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

//UpdateMeta patches the meta data right now it is used to save the UI render info (pipe/list)
func (e *Entity) UpdateMeta(ctx context.Context, db *sqlx.DB, meta map[string]interface{}) error {
	ctx, span := trace.StartSpan(ctx, "internal.entity.UpdateMeta")
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

func (e *Entity) AddTag(ctx context.Context, db *sqlx.DB, tag string, position int) error {
	ctx, span := trace.StartSpan(ctx, "internal.entity.UpdateTags")
	defer span.End()

	for _, tg := range e.Tags {
		if tg == tag {
			return nil
		}
	}

	e.Tags = insert(e.Tags, position, tag)
	log.Println("AddTag e.ID ", e.ID)
	log.Println("AddTag e.Tags ", e.Tags)
	const q = `UPDATE entities SET
		"tags" = $3 
		WHERE account_id = $1 AND entity_id = $2`
	_, err := db.ExecContext(ctx, q, e.AccountID, e.ID,
		e.Tags,
	)
	if err != nil {
		log.Println("unexpected error when adding the tags", err)
	}

	return err
}

func (e *Entity) CleanTags(ctx context.Context, db *sqlx.DB, tags []string) error {
	ctx, span := trace.StartSpan(ctx, "internal.entity.CleanTags")
	defer span.End()

	log.Println("CleanTags e.Tags ", e.Tags)
	log.Println("CleanTags tags ", tags)

	e.Tags = append(e.Tags, tags...)
	log.Println("CleanTags final similars ", e.Tags)

	const q = `UPDATE entities SET
		"tags" = $3 
		WHERE account_id = $1 AND entity_id = $2`
	_, err := db.ExecContext(ctx, q, e.AccountID, e.ID,
		e.Tags,
	)
	if err != nil {
		log.Println("unexpected error when cleaning the tags", err)
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
