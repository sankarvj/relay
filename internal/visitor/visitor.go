package visitor

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific entity is requested but does not exist.
	ErrNotFound = errors.New("Visitor not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("Visitor ID is not in its proper form")
)

func List(ctx context.Context, accountID string, page int, db *sqlx.DB) ([]Visitor, error) {
	ctx, span := trace.StartSpan(ctx, "internal.visitor.List")
	defer span.End()
	visitors := []Visitor{}

	offset := page * util.PageLimt

	const q = `SELECT * FROM visitors where account_id = $1 order by updated_at offset $2 LIMIT $3`
	if err := db.SelectContext(ctx, &visitors, q, accountID, offset, util.PageLimt); err != nil {
		return nil, errors.Wrap(err, "selecting visitors for an user")
	}
	return visitors, nil
}

func ListByEmail(ctx context.Context, accountID, email string, db *sqlx.DB) ([]Visitor, error) {
	ctx, span := trace.StartSpan(ctx, "internal.visitor.List")
	defer span.End()
	visitors := []Visitor{}

	const q = `SELECT * FROM visitors where account_id = $1 AND email = $2`
	if err := db.SelectContext(ctx, &visitors, q, accountID, email); err != nil {
		return nil, errors.Wrap(err, "selecting visitors for an user")
	}
	return visitors, nil
}

func Create(ctx context.Context, db *sqlx.DB, nv NewVisitor, now time.Time) (Visitor, error) {
	ctx, span := trace.StartSpan(ctx, "internal.visitor.Create")
	defer span.End()

	timeNow := now.UTC()
	expireAt := timeNow.AddDate(0, 0, 7)

	v := Visitor{
		VistitorID: uuid.New().String(),
		AccountID:  nv.AccountID,
		TeamID:     nv.TeamID,
		EntityID:   nv.EntityID,
		ItemID:     nv.ItemID,
		Name:       nv.Name,
		Email:      nv.Email,
		Token:      nv.Token,
		ExpireAt:   expireAt,
		CreatedAt:  timeNow,
		UpdatedAt:  timeNow.Unix(),
	}

	const q = `INSERT INTO visitors
		(visitor_id, account_id, team_id, entity_id, item_id, name, email, token, expire_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := db.ExecContext(
		ctx, q,
		v.VistitorID, v.AccountID, v.TeamID, v.EntityID, v.ItemID, v.Name, v.Email, v.Token,
		v.ExpireAt, v.CreatedAt, v.UpdatedAt,
	)
	if err != nil {
		return Visitor{}, errors.Wrap(err, "inserting visitor")
	}

	return v, nil
}

func UpdateActive(ctx context.Context, db *sqlx.DB, accountID, visitorID string, isActive bool) error {
	ctx, span := trace.StartSpan(ctx, "internal.visitor.UpdateActive")
	defer span.End()

	const q = `UPDATE visitors SET
		"active" = $3,
		"update_at" = $4
		WHERE account_id = $1 AND visitor_id = $2`
	_, err := db.ExecContext(ctx, q, accountID, visitorID,
		isActive, time.Now().UTC().Unix(),
	)
	if err != nil {
		return errors.Wrap(err, "updating visitor active state")
	}

	return nil
}

func UpdateSigned(ctx context.Context, db *sqlx.DB, accountID, visitorID string) error {
	ctx, span := trace.StartSpan(ctx, "internal.visitor.UpdateActive")
	defer span.End()

	const q = `UPDATE visitors SET
		"signed_in" = $3,
		"update_at" = $4
		WHERE account_id = $1 AND visitor_id = $2`
	_, err := db.ExecContext(ctx, q, accountID, visitorID,
		true, time.Now().UTC().Unix(),
	)
	if err != nil {
		return errors.Wrap(err, "updating visitor signed state")
	}

	return nil
}

func UpdateExpiry(ctx context.Context, db *sqlx.DB, accountID, visitorID string) error {
	ctx, span := trace.StartSpan(ctx, "internal.visitor.UpdateExpiry")
	defer span.End()

	timeNow := time.Now().UTC()
	expireAt := timeNow.AddDate(0, 0, 7)

	const q = `UPDATE visitors SET
		"expire_at" = $3,
		"update_at" = $4
		WHERE account_id = $1 AND visitor_id = $2`
	_, err := db.ExecContext(ctx, q, accountID, visitorID,
		expireAt, time.Now().UTC().Unix(),
	)
	if err != nil {
		return errors.Wrap(err, "updating visitor expiry")
	}

	return nil
}

func Retrieve(ctx context.Context, accountID, visitorID string, db *sqlx.DB) (Visitor, error) {
	ctx, span := trace.StartSpan(ctx, "internal.layout.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(visitorID); err != nil {
		return Visitor{}, ErrInvalidID
	}

	var v Visitor
	const q = `SELECT * FROM visitors WHERE account_id = $1 AND visitor_id = $2`
	if err := db.GetContext(ctx, &v, q, accountID, visitorID); err != nil {
		if err == sql.ErrNoRows {
			return Visitor{}, ErrNotFound
		}

		return Visitor{}, errors.Wrapf(err, "selecting visitor %q", visitorID)
	}

	return v, nil
}

func Delete(ctx context.Context, db *sqlx.DB, accountID, visitorID string) error {
	ctx, span := trace.StartSpan(ctx, "internal.visitor.Delete")
	defer span.End()

	const q = `DELETE FROM visitors WHERE account_id = $1 and visitor_id = $2`

	if _, err := db.ExecContext(ctx, q, accountID, visitorID); err != nil {
		return errors.Wrapf(err, "deleting visitor %s", visitorID)
	}

	return nil
}
