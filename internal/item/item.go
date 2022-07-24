package item

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific entity is requested but does not exist.
	ErrNotFound = errors.New("Item not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("Item ID is not in its proper form")
)

// List retrieves a list of existing item for the entity associated from the database.
func List(ctx context.Context, entityID string, db *sqlx.DB) ([]Item, error) {
	return ListFilterByState(ctx, entityID, StateDefault, db)
}
func ListFilterByState(ctx context.Context, entityID string, state int, db *sqlx.DB) ([]Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.List")
	defer span.End()

	items := []Item{}
	const q = `SELECT * FROM items where entity_id = $1 AND state = $2 LIMIT $3`

	if err := db.SelectContext(ctx, &items, q, entityID, state, util.MaxLimt); err != nil {
		return nil, errors.Wrap(err, "selecting items")
	}

	return items, nil
}

func BulkRetrieveItems(ctx context.Context, accountID string, ids []interface{}, db *sqlx.DB) ([]Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.BulkRetrieve")
	defer span.End()

	items := []Item{}
	const q = `SELECT * FROM items where account_id = $1 AND item_id = any($2)`

	if err := db.SelectContext(ctx, &items, q, accountID, pq.Array(ids)); err != nil {
		return items, errors.Wrap(err, "selecting bulk items for selected item ids")
	}

	return items, nil
}

func EntityItems(ctx context.Context, entityID string, db *sqlx.DB) ([]Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.EntityItems")
	defer span.End()

	items := []Item{}
	const q = `SELECT * FROM items where entity_id = $1 LIMIT 50`

	if entityID != "" {
		if err := db.SelectContext(ctx, &items, q, entityID); err != nil {
			return items, errors.Wrap(err, "selecting bulk items for entity id")
		}
	}

	return items, nil
}

func UserEntityItems(ctx context.Context, entityID, userID string, db *sqlx.DB) ([]Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.UserEntityItems")
	defer span.End()

	items := []Item{}
	const q = `SELECT * FROM items where entity_id = $1 AND user_id = $2 LIMIT 20`

	if err := db.SelectContext(ctx, &items, q, entityID, userID); err != nil {
		return items, errors.Wrap(err, "selecting bulk items for entity id with user")
	}

	return items, nil
}

//TODO add pagination
func GenieEntityItems(ctx context.Context, entityIDs []string, genieID string, db *sqlx.DB) ([]Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.GenieEntityItems")
	defer span.End()

	items := []Item{}
	const q = `SELECT * FROM items where entity_id = any($1) AND genie_id = $2 LIMIT 20`

	if err := db.SelectContext(ctx, &items, q, pq.Array(entityIDs), genieID); err != nil {
		return items, errors.Wrap(err, "selecting bulk items for entity id with genie")
	}

	return items, nil
}

// Create inserts a new item into the database.
func Create(ctx context.Context, db *sqlx.DB, n NewItem, now time.Time) (Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.Create")
	defer span.End()

	fieldsBytes, err := json.Marshal(n.Fields)
	if err != nil {
		return Item{}, errors.Wrap(err, "encode fields to bytes")
	}

	if n.GenieID != nil && *n.GenieID == "" {
		n.GenieID = nil
	}

	if n.UserID != nil && *n.UserID == "" {
		n.UserID = nil
	}

	if n.StageID != nil && *n.StageID == "" {
		n.StageID = nil
	}

	i := Item{
		ID:        n.ID,
		AccountID: n.AccountID,
		EntityID:  n.EntityID,
		GenieID:   n.GenieID,
		UserID:    n.UserID,
		StageID:   n.StageID,
		Name:      n.Name,
		Type:      n.Type,
		State:     n.State,
		Fieldsb:   string(fieldsBytes),
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC().Unix(),
	}

	const q = `INSERT INTO items
		(item_id, account_id, entity_id, genie_id, user_id, stage_id, name, type, state, fieldsb, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err = db.ExecContext(
		ctx, q,
		i.ID, i.AccountID, i.EntityID, i.GenieID, i.UserID, i.StageID, i.Name, i.Type, i.State, i.Fieldsb,
		i.CreatedAt, i.UpdatedAt,
	)
	if err != nil {
		return Item{}, errors.Wrap(err, "inserting item")
	}

	return i, nil
}

//UpdateFields patches the field data
func UpdateFields(ctx context.Context, db *sqlx.DB, entityID, id string, fields map[string]interface{}) (Item, error) {
	input, err := json.Marshal(fields)
	if err != nil {
		return Item{}, errors.Wrap(err, "encode fields to input")
	}
	inputStr := string(input)
	upd := UpdateItem{
		Fieldsb: &inputStr,
	}
	return update(ctx, db, entityID, id, upd, time.Now())
}

// Update replaces a item document in the database.
func update(ctx context.Context, db *sqlx.DB, entityID, id string, upd UpdateItem, now time.Time) (Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.Update")
	defer span.End()
	i, err := Retrieve(ctx, entityID, id, db)
	if err != nil {
		return Item{}, err
	}

	if upd.Fieldsb != nil {
		i.Fieldsb = *upd.Fieldsb
	}
	i.UpdatedAt = now.Unix()

	const q = `UPDATE items SET
		"fieldsb" = $2,
		"updated_at" = $3
		WHERE item_id = $1`
	_, err = db.ExecContext(ctx, q, i.ID,
		i.Fieldsb, i.UpdatedAt,
	)
	if err != nil {
		return Item{}, errors.Wrap(err, "updating item")
	}

	return i, nil
}

// Retrieve gets the specified user from the database.
func Retrieve(ctx context.Context, entityID, itemID string, db *sqlx.DB) (Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(itemID); err != nil {
		return Item{}, ErrInvalidID
	}

	var i Item
	const q = `SELECT * FROM items WHERE entity_id = $1 AND item_id = $2`
	if err := db.GetContext(ctx, &i, q, entityID, itemID); err != nil {
		if err == sql.ErrNoRows {
			return Item{}, ErrNotFound
		}

		return Item{}, errors.Wrapf(err, "selecting item %q", itemID)
	}

	return i, nil
}

func Delete(ctx context.Context, db *sqlx.DB, accountID, entityID, itemID string) error {
	ctx, span := trace.StartSpan(ctx, "internal.item.Delete")
	defer span.End()

	const q = `DELETE FROM items WHERE account_id = $1 and entity_id = $2 and item_id = $3`

	if _, err := db.ExecContext(ctx, q, accountID, entityID, itemID); err != nil {
		return errors.Wrapf(err, "deleting item %s", itemID)
	}

	return nil
}

func DeleteAllByGenie(ctx context.Context, db *sqlx.DB, accountID, entityID, genieID string) error {
	ctx, span := trace.StartSpan(ctx, "internal.item.DeleteAllByGenie")
	defer span.End()

	const q = `DELETE FROM items WHERE account_id = $1 and entity_id = $2 and genie_id = $3`

	if _, err := db.ExecContext(ctx, q, accountID, entityID, genieID); err != nil {
		return errors.Wrapf(err, "deleting items for account %s on entity %s", accountID, entityID)
	}

	return nil
}

func DeleteAllByUser(ctx context.Context, db *sqlx.DB, accountID, entityID, userID string) error {
	ctx, span := trace.StartSpan(ctx, "internal.item.DeleteAllByUser")
	defer span.End()

	const q = `DELETE FROM items WHERE account_id = $1 and entity_id = $2 and user_id = $3`

	if _, err := db.ExecContext(ctx, q, accountID, entityID, userID); err != nil {
		return errors.Wrapf(err, "deleting items for account %s on entity %s", accountID, entityID)
	}

	return nil
}

// Fields parses attribures to fields
func (i Item) Fields() map[string]interface{} {
	var fields map[string]interface{}
	if i.Fieldsb == "" {
		return fields
	}
	if err := json.Unmarshal([]byte(i.Fieldsb), &fields); err != nil {
		log.Printf("***> unexpected error occurred when unmarshalling fields for item: %v error: %v\n", i.ID, err)
	}
	return fields
}

//Diff old and new fields
func Diff(oldItemFields, newItemFields map[string]interface{}) map[string]interface{} {
	if oldItemFields == nil {
		return newItemFields
	}

	diffFields := make(map[string]interface{}, 0)
	for k, v := range newItemFields {
		diffFields[k] = v
	}
	for key, newItem := range newItemFields {
		if oldItem, ok := oldItemFields[key]; ok {
			if ruler.Compare(newItem, oldItem) {
				log.Printf("internal.item diff : no change detected for key %s\n", key)
				delete(diffFields, key)
			} else {
				log.Printf("internal.item diff : change captured for key %s\n", key)
			}
		}
	}
	return diffFields
}

func CompareItems(newItemVals, oldItemVals []interface{}) ([]interface{}, []interface{}) {
	var oi int
	var oldVal interface{}
	oldItems := oldItemVals
	newItems := newItemVals
	for ni, newVal := range newItemVals {
		exist := false
		for oi, oldVal = range oldItemVals {
			if newVal == oldVal {
				exist = true
				break
			}
		}
		if exist {
			removeIndex(oldItems, oi)
			removeIndex(newItems, ni)
		}
	}
	return oldItems, newItems
}

func FetchIDs(items []Item) []string {
	ids := make([]string, 0)
	for _, i := range items {
		ids = append(ids, i.ID)
	}
	return ids
}

func removeIndex(s []interface{}, index int) []interface{} {
	return append(s[:index], s[index+1:]...)
}
