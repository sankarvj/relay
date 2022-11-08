package item

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"go.opencensus.io/trace"
)

//UpdateMeta patches the meta data right now it is used to save the UI web forms
func UpdateFieldsWithMeta(ctx context.Context, db *sqlx.DB, i Item, name *string, fields, meta map[string]interface{}, isPublic bool) (Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.UpdateMeta")
	defer span.End()

	i.Name = name
	i.IsPublic = isPublic
	i.Metab = i.metabJson(meta)
	i.Fieldsb = *marshalMap(fields)
	i.UpdatedAt = time.Now().Unix()

	const q = `UPDATE items SET
		"name"  = $4,
		"metab" = $5,
		"fieldsb" = $6,
		"updated_at" = $7,
		"is_public" = $8 
		WHERE account_id = $1 AND entity_id = $2 AND item_id = $3`
	_, err := db.ExecContext(ctx, q, i.AccountID, i.EntityID, i.ID,
		i.Name, i.Metab, i.Fieldsb, i.UpdatedAt, i.IsPublic,
	)
	return i, err
}

func UpdatePublicAccess(ctx context.Context, db *sqlx.DB, i Item) (Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.UpdatePublicAccess")
	defer span.End()

	i.UpdatedAt = time.Now().Unix()

	const q = `UPDATE items SET
		"updated_at" = $4,
		"is_public" = $5 
		WHERE account_id = $1 AND entity_id = $2 AND item_id = $3`
	_, err := db.ExecContext(ctx, q, i.AccountID, i.EntityID, i.ID, i.UpdatedAt, i.IsPublic)
	return i, err
}

func (i Item) Meta() map[string]interface{} {
	meta := make(map[string]interface{}, 0)
	if i.Metab == nil || *i.Metab == "" {
		return meta
	}
	if err := json.Unmarshal([]byte(*i.Metab), &meta); err != nil {
		log.Printf("***> unexpected error occurred when unmarshalling meta for item: %v error: %v\n", i.ID, err)
	}
	return meta
}

func (i Item) metabJson(meta map[string]interface{}) *string {
	existingMeta := i.Meta()
	for key, value := range meta {
		existingMeta[key] = value
	}
	return marshalMap(existingMeta)
}

func marshalMap(meta map[string]interface{}) *string {
	input, _ := json.Marshal(meta)
	metab := string(input)
	return &metab
}
