package connection

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound occurs when the flow is not found
	ErrNotFound = errors.New("Connection not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	UUIDHOLDER = "00000000-0000-0000-0000-000000000000"
)

//Associate items
func Associate(ctx context.Context, db *sqlx.DB, accountID, userID, relationshipID, entityName, srcEntityID, dstEntityID, srcItemID, dstItemID string, valueAddedFields []entity.Field, action string) error {
	now := time.Now()

	dynamicPlaceHolder := make(map[string]interface{}, 0)
	// value add properties
	for _, vaf := range valueAddedFields {
		dynamicPlaceHolder[vaf.Meta["layout"]] = vaf.Value
	}

	var title string
	var subTitle string

	if dynamicPlaceHolder["title"] != nil {
		title = dynamicPlaceHolder["title"].(string)
	}
	if dynamicPlaceHolder["sub_title"] != nil {
		subTitle = dynamicPlaceHolder["sub_title"].(string)
	}

	c := Connection{
		ConnectionID:   uuid.New().String(), //Useful only for pagination
		AccountID:      accountID,
		UserID:         userID,
		RelationshipID: relationshipID,
		EntityName:     entityName,
		SrcEntityID:    srcEntityID,
		DstEntityID:    dstEntityID,
		SrcItemID:      srcItemID,
		DstItemID:      dstItemID,
		Title:          title,
		SubTitle:       subTitle,
		Action:         action,
		CreatedAt:      now.UTC(),
		UpdatedAt:      now.UTC().Unix(),
	}

	_, err := Create(ctx, db, c)
	return err
}

// Create add new connection with respective types.
func Create(ctx context.Context, db *sqlx.DB, c Connection) (Connection, error) {
	ctx, span := trace.StartSpan(ctx, "internal.connection.Create")
	defer span.End()

	const q = `INSERT INTO connections
		(connection_id, account_id, user_id, relationship_id, entity_name, src_entity_id, dst_entity_id, src_item_id, dst_item_id, title, sub_title, action, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := db.ExecContext(
		ctx, q, c.ConnectionID,
		c.AccountID, c.UserID, c.RelationshipID, c.EntityName, c.SrcEntityID, c.DstEntityID, c.SrcItemID, c.DstItemID, c.Title, c.SubTitle, c.Action, c.CreatedAt, c.UpdatedAt)
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

//JustChildItemIDs is just ChildItemIDs but relationshipID.
//JustChildItemIDs is called from the events API to get all the associated items in one shot.
func JustChildItemIDs(ctx context.Context, db *sqlx.DB, accountID, itemID string, next string) ([]Connection, error) {
	ctx, span := trace.StartSpan(ctx, "internal.connection.JustChildItemIDs")
	defer span.End()

	var childItems []Connection
	if next != "" {
		const q = `SELECT * FROM connections where account_id = $1 AND ( dst_item_id = $2 ) AND connection_id > $3  ORDER BY created_at DESC LIMIT 5`
		if err := db.SelectContext(ctx, &childItems, q, accountID, itemID, next); err != nil {
			return nil, errors.Wrap(err, "selecting src items for connected dst item")
		}
	} else {
		const q = `SELECT * FROM connections where account_id = $1 AND ( dst_item_id = $2 ) ORDER BY created_at DESC LIMIT 5`
		if err := db.SelectContext(ctx, &childItems, q, accountID, itemID); err != nil {
			return nil, errors.Wrap(err, "selecting src items for connected dst item")
		}
	}

	return childItems, nil
}

func Delete(ctx context.Context, db *sqlx.DB, relationshipID, dstItemID string) error {
	ctx, span := trace.StartSpan(ctx, "internal.connection.Delete")
	defer span.End()

	const q = `DELETE FROM connections WHERE relationship_id = $1 and dst_item_id =$2`

	if _, err := db.ExecContext(ctx, q, relationshipID, dstItemID); err != nil {
		return errors.Wrapf(err, "deleting connection for relationship %s on %s", relationshipID, dstItemID)
	}

	return nil
}

func pickOpposites(itemID string, childItemids []Connection) []interface{} {
	itemIds := make([]interface{}, len(childItemids))
	for i, c := range childItemids {
		if c.SrcItemID == itemID {
			itemIds[i] = c.DstItemID
		} else {
			itemIds[i] = c.SrcItemID
		}
	}
	return itemIds
}

func (c Connection) PickOpposite(itemID string) (string, string) {
	if c.SrcItemID == itemID {
		return c.DstEntityID, c.DstItemID
	}
	return c.SrcEntityID, c.SrcItemID
}
