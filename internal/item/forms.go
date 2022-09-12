package item

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"go.opencensus.io/trace"
)

type Form struct {
	ID         string `db:"item_id" json:"id"`
	EntityID   string `db:"entity_id"  json:"entity_id"`
	AccountID  string `db:"account_id"  json:"account_id"`
	TeamID     string `db:"team_id" json:"team_id"`
	Name       string `db:"name" json:"name"`
	EntityName string `db:"entity_name" json:"entity_name"`
}

func Forms(ctx context.Context, accountID string, entityIDs []string, page int, db *sqlx.DB) ([]Form, error) {
	ctx, span := trace.StartSpan(ctx, "internal.forms.List")
	defer span.End()
	forms := []Form{}

	if len(entityIDs) == 0 {
		return forms, nil
	}

	offset := page * util.PageLimt

	q, args, err := sqlx.In(`SELECT i.item_id as item_id, i.entity_id as entity_id, i.account_id, e.team_id, i.name, e.display_name as entity_name FROM items as i join entities as e on i.entity_id = e.entity_id where i.account_id = ? AND i.entity_id IN (?) AND i.state = ? order by i.updated_at offset ? LIMIT ?`, accountID, entityIDs, StateWebForm, offset, util.PageLimt)
	if err != nil {
		return nil, errors.Wrap(err, "selecting forms")
	}
	q = db.Rebind(q)
	if err := db.SelectContext(ctx, &forms, q, args...); err != nil {
		return nil, errors.Wrap(err, "selecting forms")
	}

	return forms, nil
}

func Count(ctx context.Context, accountID string, entityIDs []string, db *sqlx.DB) (int, error) {
	ctx, span := trace.StartSpan(ctx, "internal.forms.Count")
	defer span.End()
	type Counter struct {
		Count int `json:"count"`
	}
	//not an ideal way....
	counts := []Counter{}

	if len(entityIDs) == 0 {
		return 0, nil
	}

	q, args, err := sqlx.In(`SELECT count(*) FROM items where account_id = ? AND entity_id IN (?) AND state = ?`, accountID, entityIDs, StateWebForm)
	if err != nil {
		return 0, errors.Wrap(err, "counting total forms")
	}
	q = db.Rebind(q)
	if err := db.SelectContext(ctx, &counts, q, args...); err != nil {
		return 0, errors.Wrap(err, "counting total forms")
	}

	if len(counts) > 0 {
		return counts[0].Count, nil
	}

	return 0, nil
}
