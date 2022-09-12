package item

import (
	"context"
	"encoding/json"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

//UpdateMeta patches the meta data right now it is used to save the UI web forms
func (i *Item) UpdateMeta(ctx context.Context, db *sqlx.DB, name *string, meta map[string]interface{}) error {
	ctx, span := trace.StartSpan(ctx, "internal.item.UpdateMeta")
	defer span.End()

	existingMeta := i.Meta()
	for key, value := range meta {
		existingMeta[key] = value
	}
	input, err := json.Marshal(existingMeta)
	if err != nil {
		return errors.Wrap(err, "encode meta to input")
	}
	metab := string(input)
	i.Name = name
	i.Metab = &metab

	const q = `UPDATE items SET
		"metab" = $4,
		"name"  = $5
		WHERE account_id = $1 AND entity_id = $2 AND item_id = $3`
	_, err = db.ExecContext(ctx, q, i.AccountID, i.EntityID, i.ID,
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
