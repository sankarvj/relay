package entity

import (
	"context"
	"encoding/json"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

const (
	MetaRender = "render" // pipe/list
)

const (
	MetaRenderPipe = "pipe" // pipe/list
	MetaRenderList = "list" // pipe/list
)

//UpdateMeta patches the meta data
func (e *Entity) UpdateMeta(ctx context.Context, db *sqlx.DB, meta map[string]interface{}) error {
	ctx, span := trace.StartSpan(ctx, "internal.item.UpdateMeta")
	defer span.End()

	existingMeta := e.Meta()
	for key, value := range meta {
		existingMeta[key] = value
	}
	input, err := json.Marshal(existingMeta)
	if err != nil {
		return errors.Wrap(err, "encode meta to input")
	}
	metab := string(input)
	e.Metab = &metab

	const q = `UPDATE entities SET
		"metab" = $3 
		WHERE account_id = $1 AND entity_id = $2`
	_, err = db.ExecContext(ctx, q, e.AccountID, e.ID,
		e.Metab,
	)
	return err
}

func (e Entity) Meta() map[string]interface{} {
	meta := make(map[string]interface{}, 0)
	if e.Metab == nil || *e.Metab == "" {
		return meta
	}
	if err := json.Unmarshal([]byte(*e.Metab), &meta); err != nil {
		log.Printf("***> unexpected error occurred when unmarshalling meta for entity: %v error: %v\n", e.ID, err)
	}
	return meta
}

func (e Entity) IsPipeLayout() bool {
	metaHash := e.Meta()
	if val, ok := metaHash[MetaRender]; ok {
		if val == "pipe" {
			return true
		}
	}
	return false
}
