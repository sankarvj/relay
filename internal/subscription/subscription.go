package subscription

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

var (
	// ErrSubscriptionNotFound is used when a specific subscription is requested but does not exist.
	ErrSubscriptionNotFound = errors.New("Subscription not found")
)

func Create(ctx context.Context, db *sqlx.DB, ns NewSubscription, now time.Time) (Subscription, error) {
	ctx, span := trace.StartSpan(ctx, "internal.subscription.Create")
	defer span.End()

	s := Subscription{
		ID:        ns.ID,
		AccountID: ns.AccountID,
		EntityID:  ns.EntityID,
		ItemID:    ns.ItemID,
		UserID:    ns.UserID,
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC().Unix(),
	}

	const q = `INSERT INTO subscriptions
		(subscription_id, account_id, entity_id, item_id, user_id,
		created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := db.ExecContext(
		ctx, q,
		s.ID, s.AccountID, s.EntityID, s.ItemID, s.UserID,
		s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return Subscription{}, errors.Wrap(err, "inserting subscription")
	}

	return s, nil
}

// Retrieve gets the specified subscription from the database.
// subscriptions are unique to the whole product. It does not wrapped inside account/entity
func Retrieve(ctx context.Context, id string, db *sqlx.DB) (*Subscription, error) {
	ctx, span := trace.StartSpan(ctx, "internal.subscription.Retrieve")
	defer span.End()

	var s Subscription
	const q = `SELECT * FROM subscriptions WHERE subscription_id = $1`
	if err := db.GetContext(ctx, &s, q, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSubscriptionNotFound
		}

		return nil, errors.Wrapf(err, "selecting subscription %q", id)
	}

	return &s, nil
}
