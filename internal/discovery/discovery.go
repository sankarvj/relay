package discovery

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

var (
	// ErrDiscoveryEmpty is used when a specific discovery is requested but does not exist.
	ErrDiscoveryEmpty = errors.New("No Discoveries found")
)

func Create(ctx context.Context, db *sqlx.DB, ns NewDiscovery, now time.Time) (Discover, error) {
	ctx, span := trace.StartSpan(ctx, "internal.discovery.Create")
	defer span.End()

	s := Discover{
		ID:        ns.ID,
		Type:      ns.Type,
		AccountID: ns.AccountID,
		EntityID:  ns.EntityID,
		ItemID:    ns.ItemID,
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC().Unix(),
	}

	const q = `INSERT INTO discoveries
		(discovery_id, discovery_type, account_id, entity_id, item_id,
		created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := db.ExecContext(
		ctx, q,
		s.ID, s.Type, s.AccountID, s.EntityID, s.ItemID,
		s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return Discover{}, errors.Wrap(err, "inserting into discoveries")
	}

	return s, nil
}

// Retrieve gets the specified discovery from the database.
// discoveries are unique to the whole product. It does not wrapped inside account/entity
func Retrieve(ctx context.Context, accountID, entityID, id string, db *sqlx.DB) (*Discover, error) {
	ctx, span := trace.StartSpan(ctx, "internal.discovery.Retrieve")
	defer span.End()

	//email integration is one-time. If integrated once he will not be able to add it again in other accounts
	//in that case we might have the accountID and entityID as 0
	if accountID == "" && entityID == "" {
		var s Discover
		const q = `SELECT * FROM discoveries WHERE discovery_id = $1`
		if err := db.GetContext(ctx, &s, q, id); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrDiscoveryEmpty
			}

			return nil, errors.Wrapf(err, "discovering id %q", id)
		}

		return &s, nil
	}

	var s Discover
	const q = `SELECT * FROM discoveries WHERE discovery_id = $1 AND account_id = $2 AND entity_id = $3`
	if err := db.GetContext(ctx, &s, q, id, accountID, entityID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrDiscoveryEmpty
		}

		return nil, errors.Wrapf(err, "discovering id %q, account_id %q, entity_id %q ", id, accountID, entityID)
	}

	return &s, nil
}
