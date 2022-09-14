package item

import (
	"context"
	"encoding/json"
	"log"

	"github.com/jmoiron/sqlx"
	"go.opencensus.io/trace"
)

//UpdateMeta patches the meta data right now it is used to save the UI web forms
func (i *Item) UpdateMeta(ctx context.Context, db *sqlx.DB, name *string, meta map[string]interface{}) error {
	ctx, span := trace.StartSpan(ctx, "internal.item.UpdateMeta")
	defer span.End()

	i.Name = name
	i.Metab = i.metabJson(meta)

	const q = `UPDATE items SET
		"metab" = $4,
		"name"  = $5
		WHERE account_id = $1 AND entity_id = $2 AND item_id = $3`
	_, err := db.ExecContext(ctx, q, i.AccountID, i.EntityID, i.ID,
		i.Metab, i.Name,
	)
	return err
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
	return marshalMeta(existingMeta)
}

func marshalMeta(meta map[string]interface{}) *string {
	input, _ := json.Marshal(meta)
	metab := string(input)
	return &metab
}
