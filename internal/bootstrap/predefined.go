package bootstrap

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific entity is requested but does not exist.
	ErrNotFound = errors.New("Entity not found")
)

func RetrievePreDefinedEntity(ctx context.Context, db *sqlx.DB, accountID string, systemEntityName string) (entity.Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.predefined.RetrieveUserEntity")
	defer span.End()

	var e entity.Entity
	const q = `SELECT * FROM entities WHERE account_id = $1 AND name = $2 LIMIT 1`
	if err := db.GetContext(ctx, &e, q, accountID, systemEntityName); err != nil {
		if err == sql.ErrNoRows {
			return entity.Entity{}, ErrNotFound
		}

		return entity.Entity{}, errors.Wrapf(err, "selecting system entity %q", systemEntityName)
	}

	return e, nil
}
