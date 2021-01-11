package item

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"

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
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// List retrieves a list of existing item for the entity associated from the database.
func List(ctx context.Context, entityID string, db *sqlx.DB) ([]Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.List")
	defer span.End()

	items := []Item{}
	const q = `SELECT * FROM items where entity_id = $1`

	if err := db.SelectContext(ctx, &items, q, entityID); err != nil {
		return nil, errors.Wrap(err, "selecting items")
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

	i := Item{
		ID:        n.ID,
		AccountID: n.AccountID,
		EntityID:  n.EntityID,
		Fieldsb:   string(fieldsBytes),
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC().Unix(),
	}

	const q = `INSERT INTO items
		(item_id, account_id, entity_id, fieldsb, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = db.ExecContext(
		ctx, q,
		i.ID, i.AccountID, i.EntityID, i.Fieldsb,
		i.CreatedAt, i.UpdatedAt,
	)
	if err != nil {
		return Item{}, errors.Wrap(err, "inserting item")
	}

	return i, nil
}

//UpdateFields patches the field data
func UpdateFields(ctx context.Context, db *sqlx.DB, entityID, id string, fields map[string]interface{}) error {
	input, err := json.Marshal(fields)
	if err != nil {
		return errors.Wrap(err, "encode fields to input")
	}
	inputStr := string(input)
	upd := UpdateItem{
		Fieldsb: &inputStr,
	}
	return update(ctx, db, entityID, id, upd, time.Now())
}

// Update replaces a item document in the database.
func update(ctx context.Context, db *sqlx.DB, entityID, id string, upd UpdateItem, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.item.Update")
	defer span.End()
	i, err := Retrieve(ctx, entityID, id, db)
	if err != nil {
		return err
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
		return errors.Wrap(err, "updating item")
	}

	return nil
}

// Retrieve gets the specified user from the database.
func Retrieve(ctx context.Context, entityID, id string, db *sqlx.DB) (Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return Item{}, ErrInvalidID
	}

	var i Item
	const q = `SELECT * FROM items WHERE entity_id = $1 AND item_id = $2`
	if err := db.GetContext(ctx, &i, q, entityID, id); err != nil {
		if err == sql.ErrNoRows {
			return Item{}, ErrNotFound
		}

		return Item{}, errors.Wrapf(err, "selecting item %q", id)
	}

	return i, nil
}

func BulkRetrieve(ctx context.Context, entityID string, ids []interface{}, db *sqlx.DB) ([]Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.BulkRetrieve")
	defer span.End()

	items := []Item{}
	const q = `SELECT * FROM items where entity_id = $1 AND item_id = any($2)`

	if err := db.SelectContext(ctx, &items, q, entityID, pq.Array(ids)); err != nil {
		return items, errors.Wrap(err, "selecting bulk items for entity id and selected item ids")
	}

	return items, nil
}

func EntityItems(ctx context.Context, entityID string, db *sqlx.DB) ([]Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.EntityItems")
	defer span.End()

	items := []Item{}
	const q = `SELECT * FROM items where entity_id = $1 LIMIT 20`

	if err := db.SelectContext(ctx, &items, q, entityID); err != nil {
		return items, errors.Wrap(err, "selecting bulk items for entity id")
	}

	return items, nil
}

// Fields parses attribures to fields
func (i Item) Fields() map[string]interface{} {
	var fields map[string]interface{}
	if err := json.Unmarshal([]byte(i.Fieldsb), &fields); err != nil {
		log.Printf("error while unmarshalling item fieldsb %v", i.ID)
		log.Println(err)
	}
	return fields
}

//Diff old and new fields
func Diff(oldItemFields, newItemFields map[string]interface{}) map[string]interface{} {
	diffFields := newItemFields
	for key, newItem := range newItemFields {
		if oldItem, ok := oldItemFields[key]; ok {
			if ruler.Compare(newItem, oldItem) {
				log.Printf("-> no change for key %s", key)
				delete(diffFields, key)
			} else {
				log.Printf("->> change captured for key %s !", key)
			}
		}
	}
	return diffFields
}

func CompareItems(oldItemVals, newItemVals []interface{}) ([]interface{}, []interface{}) {
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

func removeIndex(s []interface{}, index int) []interface{} {
	return append(s[:index], s[index+1:]...)
}
