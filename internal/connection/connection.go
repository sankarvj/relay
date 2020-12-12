package connection

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound occurs when the flow is not found
	ErrNotFound = errors.New("Connection not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// Create add new connection with respective types.
func Create(ctx context.Context, db *sqlx.DB, c Connection) (Connection, error) {
	ctx, span := trace.StartSpan(ctx, "internal.connection.Create")
	defer span.End()

	const q = `INSERT INTO connections
		(account_id, relationship_id, src_item_id, dst_item_id)
		VALUES ($1, $2, $3, $4)`

	_, err := db.ExecContext(
		ctx, q,
		c.AccountID, c.RelationshipID, c.SrcItemID, c.DstItemID)
	if err != nil {
		return Connection{}, errors.Wrap(err, "inserting connections")
	}

	return c, nil
}

func Update(ctx context.Context, db *sqlx.DB, c Connection) error {
	ctx, span := trace.StartSpan(ctx, "internal.connection.Update")
	defer span.End()

	const q = `UPDATE connections SET
		"dst_item_id" = $1
		WHERE account_id = $2 AND relationship_id = $3 AND src_item_id = $4`
	_, err := db.ExecContext(ctx, q, c.DstItemID, c.AccountID,
		c.RelationshipID, c.SrcItemID,
	)
	if err != nil {
		return errors.Wrap(err, "updating connections")
	}

	return nil
}

// List gets the connections for the destination entity
func List(ctx context.Context, db *sqlx.DB, accountID, relationshipID string) ([]Connection, error) {
	ctx, span := trace.StartSpan(ctx, "internal.connection.List")
	defer span.End()

	var connections []Connection
	const q = `SELECT * FROM connections where account_id = $1 AND relationship_id = $2`

	if err := db.SelectContext(ctx, &connections, q, accountID, relationshipID); err != nil {
		return nil, errors.Wrap(err, "selecting connections for relationship")
	}

	return connections, nil
}

func ChildItemIDs(ctx context.Context, db *sqlx.DB, accountID, relationshipID, itemID string) ([]string, error) {
	ctx, span := trace.StartSpan(ctx, "internal.connection.ChildItemIDs")
	defer span.End()

	var childItemids []string
	const q = `SELECT src_item_id FROM connections where account_id = $1 AND relationship_id = $2 AND $3 = ANY(dst_item_id)`

	if err := db.SelectContext(ctx, &childItemids, q, accountID, relationshipID, itemID); err != nil {
		return nil, errors.Wrap(err, "selecting src items for connected dst item")
	}

	return childItemids, nil
}
