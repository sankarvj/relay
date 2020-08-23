package entity

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific entity is requested but does not exist.
	ErrNotFound = errors.New("Entity not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// List retrieves a list of existing entities for the team associated from the database.
func List(ctx context.Context, teamID string, db *sqlx.DB) ([]Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.entity.List")
	defer span.End()

	entities := []Entity{}
	const q = `SELECT * FROM entities where team_id = $1`

	if err := db.SelectContext(ctx, &entities, q, teamID); err != nil {
		return nil, errors.Wrap(err, "selecting entities")
	}

	return entities, nil
}

// Create inserts a new user into the database.
func Create(ctx context.Context, db *sqlx.DB, n NewEntity, now time.Time) (Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.entity.Create")
	defer span.End()

	fieldsBytes, err := json.Marshal(n.Fields)
	if err != nil {
		return Entity{}, errors.Wrap(err, "encode fields to bytes")
	}

	e := Entity{
		ID:        uuid.New().String(),
		AccountID: n.AccountID,
		TeamID:    n.TeamID,
		Name:      n.Name,
		Category:  n.Category,
		State:     n.State,
		Fieldsb:   string(fieldsBytes),
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC().Unix(),
	}

	const q = `INSERT INTO entities
		(entity_id, account_id, team_id, name, category, state, fieldsb, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = db.ExecContext(
		ctx, q,
		e.ID, e.AccountID, e.TeamID, e.Name, e.Category, e.State, e.Fieldsb,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return Entity{}, errors.Wrap(err, "inserting entity")
	}

	return e, nil
}

// Retrieve gets the specified entity from the database.
func Retrieve(ctx context.Context, id string, db *sqlx.DB) (Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.entity.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return Entity{}, ErrInvalidID
	}

	var e Entity
	const q = `SELECT * FROM entities WHERE entity_id = $1`
	if err := db.GetContext(ctx, &e, q, id); err != nil {
		if err == sql.ErrNoRows {
			return Entity{}, ErrNotFound
		}

		return Entity{}, errors.Wrapf(err, "selecting entity %q", id)
	}

	return e, nil
}

//ParseEmailEntity creates the email entity from the field map provided
func ParseEmailEntity(params map[string]interface{}) (EmailEntity, error) {
	var eme EmailEntity
	jsonbody, err := json.Marshal(params)
	if err != nil {
		return eme, err
	}
	err = json.Unmarshal(jsonbody, &eme)
	return eme, err
}

//ParseDelayEntity creates the delay entity from the field map provided
func ParseDelayEntity(params map[string]interface{}) (DelayEntity, error) {
	var de DelayEntity
	jsonbody, err := json.Marshal(params)
	if err != nil {
		return de, err
	}
	err = json.Unmarshal(jsonbody, &de)
	return de, err
}

//ParseHookEntity creates the hook entity from the field map provided
func ParseHookEntity(params map[string]interface{}) (WebHookEntity, error) {
	var whe WebHookEntity
	jsonbody, err := json.Marshal(params)
	if err != nil {
		return whe, err
	}
	err = json.Unmarshal(jsonbody, &whe)
	return whe, err
}

// FillFieldValues updates the
func FillFieldValues(entityFields []Field, itemFields map[string]interface{}) []Field {
	updatedFields := make([]Field, 0)
	for _, field := range entityFields {
		if val, ok := itemFields[field.Key]; ok && !field.Config {
			field.Value = val
		}
		updatedFields = append(updatedFields, field)
	}
	return updatedFields
}

// Fields parses attribures to fields
func (e Entity) Fields() ([]Field, error) {
	fields, err := e.AllFields()
	if err != nil {
		return nil, err
	}
	//remove all config fields
	temp := fields[:0]
	for _, field := range fields {
		if !field.Config {
			temp = append(temp, field)
		}
	}
	fields = temp
	return fields, nil
}

// AllFields parses attribures to fields
func (e Entity) AllFields() ([]Field, error) {
	var fields []Field
	if err := json.Unmarshal([]byte(e.Fieldsb), &fields); err != nil {
		return nil, errors.Wrapf(err, "error while unmarshalling entity attributes to fields type %q", e.ID)
	}
	return fields, nil
}

func (f Field) IsKeyId() bool {
	return f.Key == FieldIdKey
}
