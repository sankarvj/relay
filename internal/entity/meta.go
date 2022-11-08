package entity

import (
	"context"
	"encoding/json"
	"log"

	"github.com/jmoiron/sqlx"
	"go.opencensus.io/trace"
)

const (
	MetaRender = "render" // pipe/list
)

const (
	MetaRenderPipe  = "pipe"  // pipe/list/group
	MetaRenderList  = "list"  // pipe/list/group
	MetaRenderGroup = "group" // pipe/list/group
)

//UpdateMeta patches the meta data right now it is used to save the UI render info (pipe/list)
func (e *Entity) UpdateMeta(ctx context.Context, db *sqlx.DB) error {
	ctx, span := trace.StartSpan(ctx, "internal.entity.UpdateMeta")
	defer span.End()

	const q = `UPDATE entities SET
		"metab" = $3 
		WHERE account_id = $1 AND entity_id = $2`
	_, err := db.ExecContext(ctx, q, e.AccountID, e.ID,
		e.Metab,
	)
	return err
}

func (e *Entity) UpdatePublicAccess(ctx context.Context, db *sqlx.DB) error {
	ctx, span := trace.StartSpan(ctx, "internal.entity.UpdatePublicAccess")
	defer span.End()

	const q = `UPDATE entities SET
		"is_public" = $3 
		WHERE account_id = $1 AND entity_id = $2`
	_, err := db.ExecContext(ctx, q, e.AccountID, e.ID,
		e.IsPublic,
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

func (e Entity) Layout() string {
	metaHash := e.Meta()
	if val, ok := metaHash[MetaRender]; ok {
		return val.(string)
	}
	return MetaRenderList
}

func (e *Entity) AddTag(ctx context.Context, db *sqlx.DB, tag string, position int) error {
	ctx, span := trace.StartSpan(ctx, "internal.entity.UpdateTags")
	defer span.End()

	for _, tg := range e.Tags {
		if tg == tag {
			return nil
		}
	}

	e.Tags = insert(e.Tags, position, tag)
	const q = `UPDATE entities SET
		"tags" = $3 
		WHERE account_id = $1 AND entity_id = $2`
	_, err := db.ExecContext(ctx, q, e.AccountID, e.ID,
		e.Tags,
	)
	if err != nil {
		log.Println("***> unexpected error when adding the tags", err)
	}

	return err
}

func (e *Entity) CleanTags(ctx context.Context, db *sqlx.DB, tags []string) error {
	ctx, span := trace.StartSpan(ctx, "internal.entity.CleanTags")
	defer span.End()

	e.Tags = append(e.Tags, tags...)
	const q = `UPDATE entities SET
		"tags" = $3 
		WHERE account_id = $1 AND entity_id = $2`
	_, err := db.ExecContext(ctx, q, e.AccountID, e.ID,
		e.Tags,
	)
	if err != nil {
		log.Println("***> unexpected error when cleaning the tags", err)
	}
	return err
}

func insert(a []string, index int, value string) []string {
	if a == nil {
		a = []string{}
	}

	if len(a) == 0 || len(a) == index { // nil or empty slice or after last element
		return append(a, value)
	}
	a = append(a[:index+1], a[index:]...) // index < len(a)
	a[index] = value
	return a
}
