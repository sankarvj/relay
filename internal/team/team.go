package team

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// List retrieves a list of existing teams for the account associated from the database.
func List(ctx context.Context, accountID string, db *sqlx.DB) ([]Team, error) {
	ctx, span := trace.StartSpan(ctx, "internal.team.List")
	defer span.End()

	teams := []Team{}
	const q = `SELECT * FROM teams where account_id = $1`

	if err := db.SelectContext(ctx, &teams, q, accountID); err != nil {
		return nil, errors.Wrap(err, "selecting teams")
	}
	return teams, nil
}

// Create inserts a new user into the database.
func Create(ctx context.Context, db *sqlx.DB, n NewTeam, now time.Time) (*Team, error) {
	ctx, span := trace.StartSpan(ctx, "internal.team.Create")
	defer span.End()

	t := Team{
		Name:      n.Name,
		AccountID: n.AccountID,
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC().Unix(),
	}

	const q = `INSERT INTO teams
		(account_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)`
	_, err := db.ExecContext(
		ctx, q,
		t.AccountID, t.Name,
		t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "inserting team")
	}

	return &t, nil
}
