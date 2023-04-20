package item

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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
func List(ctx context.Context, accountID, entityID string, db *sqlx.DB) ([]Item, error) {
	return ListFilterByState(ctx, accountID, entityID, StateDefault, db)
}
func ListFilterByState(ctx context.Context, accountID, entityID string, state int, db *sqlx.DB) ([]Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.List")
	defer span.End()

	items := []Item{}
	const q = `SELECT * FROM items where account_id = $1 AND entity_id = $2 AND state = $3 LIMIT $4`

	if err := db.SelectContext(ctx, &items, q, accountID, entityID, state, util.MaxLimt); err != nil {
		return nil, errors.Wrap(err, "selecting items by state")
	}

	return items, nil
}

func Result(ctx context.Context, accountID, entityID string, pageNo int, wh string, db *sqlx.DB) ([]Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.Result")
	defer span.End()

	var pageLimt int = util.PageLimt
	var skipCount int = 0
	if pageNo == -1 {
		pageLimt = 1000
	} else {
		skipCount = pageNo * util.PageLimt
	}

	items := []Item{}
	q := fmt.Sprintf(`SELECT * FROM items where account_id = $1 AND entity_id = $2 AND state = $3 %s ORDER BY created_at DESC LIMIT $4 OFFSET $5`, wh)
	if err := db.SelectContext(ctx, &items, q, accountID, entityID, StateDefault, pageLimt, skipCount); err != nil {
		return nil, errors.Wrap(err, "selecting items result")
	}

	return items, nil
}

func Counts(ctx context.Context, accountID, entityID string, wh string, db *sqlx.DB) (map[string]int, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.Counts")
	defer span.End()

	var count int
	q := fmt.Sprintf(`SELECT count(*) as total_count FROM items where account_id = $1 AND entity_id = $2 AND state = $3 %s`, wh)

	if err := db.GetContext(ctx, &count, q, accountID, entityID, StateDefault); err != nil {
		return nil, errors.Wrap(err, "selecting items count")
	}

	return map[string]int{"total_count": count}, nil
}

func CountMap(ctx context.Context, accountID, entityID string, se, wh, grp string, db *sqlx.DB) ([]Counter, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.CountMap")
	defer span.End()

	counters := []Counter{}
	q := fmt.Sprintf(`SELECT %s FROM items where account_id = $1 AND entity_id = $2 AND state = $3 %s %s`, se, wh, grp)
	//log.Printf("CountMap q---------------------:::: %+v", q)
	if err := db.SelectContext(ctx, &counters, q, accountID, entityID, StateDefault); err != nil {
		return nil, errors.Wrap(err, "selecting items count map")
	}

	return counters, nil
}

func BulkRetrieveItems(ctx context.Context, accountID string, ids []interface{}, db *sqlx.DB) ([]Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.BulkRetrieve")
	defer span.End()

	items := []Item{}
	const q = `SELECT * FROM items where account_id = $1 AND item_id = any($2) ORDER BY created_at DESC`

	if err := db.SelectContext(ctx, &items, q, accountID, pq.Array(ids)); err != nil {
		return items, errors.Wrap(err, "selecting items for a list of item ids")
	}

	return items, nil
}

func EntityItems(ctx context.Context, accountID, entityID string, db *sqlx.DB) ([]Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.EntityItems")
	defer span.End()

	items := []Item{}
	const q = `SELECT * FROM items where account_id = $1 AND entity_id = $2 LIMIT 50`

	if entityID != "" {
		if err := db.SelectContext(ctx, &items, q, accountID, entityID); err != nil {
			return items, errors.Wrap(err, "selecting items for an entity")
		}
	}

	return items, nil
}

func TaskItems(ctx context.Context, accountID, entityID, itemID, taskEntityID string, db *sqlx.DB) ([]Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.TaskItems")
	defer span.End()

	genieID := fmt.Sprintf("%s#%s", entityID, itemID)

	items := []Item{}
	const q = `SELECT * FROM items where account_id = $1 AND entity_id = $2 AND genie_id = $3 AND LIMIT 1000`

	if err := db.SelectContext(ctx, &items, q, accountID, taskEntityID, genieID); err != nil {
		return items, errors.Wrap(err, "selecting items for an entity")
	}

	return items, nil
}

func UserEntityItems(ctx context.Context, accountID, entityID, userID string, db *sqlx.DB) ([]Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.UserEntityItems")
	defer span.End()

	items := []Item{}
	const q = `SELECT * FROM items where account_id = $1 AND entity_id = $2 AND user_id = $3 LIMIT 20`

	if err := db.SelectContext(ctx, &items, q, accountID, entityID, userID); err != nil {
		return items, errors.Wrap(err, "selecting items for user")
	}

	return items, nil
}

// Create inserts a new item into the database.
func Create(ctx context.Context, db *sqlx.DB, n NewItem, now time.Time) (Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.Create")
	defer span.End()

	//convert empty string to null
	for k, v := range n.Fields {

		if v == "" {
			n.Fields[k] = nil
		}

	}

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

	i := Item{
		ID:        n.ID,
		AccountID: n.AccountID,
		EntityID:  n.EntityID,
		GenieID:   n.GenieID,
		UserID:    n.UserID,
		Name:      n.Name,
		Type:      n.Type,
		State:     n.State,
		IsPublic:  n.IsPublic,
		Fieldsb:   string(fieldsBytes),
		Metab:     marshalMap(n.Meta),
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC().Unix(),
	}

	const q = `INSERT INTO items
		(item_id, account_id, entity_id, genie_id, user_id, name, type, state, is_public, fieldsb, metab, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err = db.ExecContext(
		ctx, q,
		i.ID, i.AccountID, i.EntityID, i.GenieID, i.UserID, i.Name, i.Type, i.State, i.IsPublic, i.Fieldsb, i.Metab,
		i.CreatedAt, i.UpdatedAt,
	)
	if err != nil {
		return Item{}, errors.Wrap(err, "inserting item")
	}

	return i, nil
}

// UpdateFields patches the field data
func UpdateFields(ctx context.Context, db *sqlx.DB, accountID, entityID, id string, fields map[string]interface{}) (Item, error) {
	//convert empty string to null
	for k, v := range fields {
		if v == "" {
			fields[k] = nil
		}
	}
	input, err := json.Marshal(fields)
	if err != nil {
		return Item{}, errors.Wrap(err, "encode fields to input")
	}
	inputStr := string(input)
	upd := UpdateItem{
		Fieldsb: &inputStr,
	}
	return update(ctx, db, accountID, entityID, id, upd, time.Now())
}

// Update replaces a item document in the database.
func update(ctx context.Context, db *sqlx.DB, accountID, entityID, id string, upd UpdateItem, now time.Time) (Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.Update")
	defer span.End()
	i, err := Retrieve(ctx, accountID, entityID, id, db)
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
func Retrieve(ctx context.Context, accountID, entityID, itemID string, db *sqlx.DB) (Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(itemID); err != nil {
		return Item{}, ErrInvalidID
	}

	var i Item
	const q = `SELECT * FROM items WHERE account_id = $1 AND entity_id = $2 AND item_id = $3`
	if err := db.GetContext(ctx, &i, q, accountID, entityID, itemID); err != nil {
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

func DeleteAllByUser(ctx context.Context, db *sqlx.DB, accountID, entityID, userID string) error {
	ctx, span := trace.StartSpan(ctx, "internal.item.DeleteAllByUser")
	defer span.End()

	const q = `DELETE FROM items WHERE account_id = $1 and entity_id = $2 and user_id = $3`

	if _, err := db.ExecContext(ctx, q, accountID, entityID, userID); err != nil {
		return errors.Wrapf(err, "deleting items for account %s on entity %s", accountID, entityID)
	}

	return nil
}

func DeleteAllByDummies(ctx context.Context, db *sqlx.DB, accountID string) error {
	ctx, span := trace.StartSpan(ctx, "internal.item.DeleteAllByDummies")
	defer span.End()

	const q = `DELETE FROM items WHERE account_id = $1 and type = $2`

	if _, err := db.ExecContext(ctx, q, accountID, TypeDummy); err != nil {
		return errors.Wrapf(err, "deleting all dummy items for account %s", accountID)
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

// Diff old and new fields
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
				//log.Printf("internal.item diff : no change detected for key %s\n", key)
				delete(diffFields, key)
			} else {
				//log.Printf("internal.item diff : change captured for key %s\n", key)
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
