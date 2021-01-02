package entity

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific entity is requested but does not exist.
	ErrNotFound = errors.New("Entity not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// List retrieves a list of existing entities for the team associated from the database.
func List(ctx context.Context, teamID string, categoryIds []int, db *sqlx.DB) ([]Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.entity.List")
	defer span.End()
	entities := []Entity{}
	if len(categoryIds) == 0 {
		const q = `SELECT * FROM entities where team_id = $1`
		if err := db.SelectContext(ctx, &entities, q, teamID); err != nil {
			return nil, errors.Wrap(err, "selecting entities for all category")
		}
	} else {
		const q = `SELECT * FROM entities where team_id = $1 AND category = any($2)`
		if err := db.SelectContext(ctx, &entities, q, teamID, pq.Array(categoryIds)); err != nil {
			return nil, errors.Wrap(err, "selecting entities for category")
		}
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
		ID:          n.ID,
		AccountID:   n.AccountID,
		TeamID:      n.TeamID,
		Name:        n.Name,
		DisplayName: n.DisplayName,
		Category:    n.Category,
		State:       n.State,
		Fieldsb:     string(fieldsBytes),
		CreatedAt:   now.UTC(),
		UpdatedAt:   now.UTC().Unix(),
	}

	const q = `INSERT INTO entities
		(entity_id, account_id, team_id, name, display_name, category, state, fieldsb, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err = db.ExecContext(
		ctx, q,
		e.ID, e.AccountID, e.TeamID, e.Name, e.DisplayName, e.Category, e.State, e.Fieldsb,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return Entity{}, errors.Wrap(err, "inserting entity")
	}

	//TODO: do it in the same transaction.
	//TODO: this relationship should happen only if the user explicitly specifies that.
	//may be, we can give add the boolean in the meta to identify that.
	relationships := populateBonds(e.AccountID, e.ID, n.Fields)
	for _, r := range relationships {
		_, err := relationship.Create(ctx, db, r)
		if err != nil {
			return e, errors.Wrapf(err, "Relationship for entity failed: %+v", e.ID)
		}
	}

	return e, nil
}

// Update replaces a item document in the database.
func Update(ctx context.Context, db *sqlx.DB, entityID string, fieldsB string, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.item.Update")
	defer span.End()

	e, err := Retrieve(ctx, entityID, db)
	if err != nil {
		return err
	}

	e.UpdatedAt = now.Unix()
	e.Fieldsb = fieldsB

	const q = `UPDATE entities SET
		"fieldsb" = $2,
		"updated_at" = $3
		WHERE entity_id = $1`
	_, err = db.ExecContext(ctx, q, e.ID,
		e.Fieldsb, e.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

//Associate entities
func Associate(ctx context.Context, db *sqlx.DB, accountID, srcEntityID, dstEntityID string) (string, error) {
	relationshipID, relationships := populateAssociation(accountID, srcEntityID, dstEntityID)
	//TODO batch create
	for _, r := range relationships {
		_, err := relationship.Create(ctx, db, r)
		if err != nil {
			return relationshipID, errors.Wrapf(err, "Association between entities %s and %s failed", srcEntityID, dstEntityID)
		}
	}
	return relationshipID, nil
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

func BulkRetrieve(ctx context.Context, ids []string, db *sqlx.DB) ([]Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.entity.BulkRetrieve")
	defer span.End()

	entities := []Entity{}
	const q = `SELECT * FROM entities where entity_id = any($1)`

	if err := db.SelectContext(ctx, &entities, q, pq.Array(ids)); err != nil {
		return entities, errors.Wrap(err, "selecting bulk entities")
	}

	return entities, nil
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
		if val, ok := itemFields[field.Key]; ok && !field.isConfig() {
			field.Value = val
		}
		updatedFields = append(updatedFields, field)
	}
	return updatedFields
}

func FetchIDs(entities []Entity) []string {
	ids := make([]string, 0)
	for _, e := range entities {
		ids = append(ids, e.ID)
	}
	return ids
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
		if !field.isConfig() {
			temp = append(temp, field)
		}
	}
	fields = temp
	return fields, nil
}

func (e Entity) FieldsWithReference() ([]*Field, error) {
	var referencedFields []*Field
	fields, err := e.Fields()
	if err != nil {
		return referencedFields, err
	}
	for i := 0; i < len(fields); i++ {
		referencedFields = append(referencedFields, &fields[i])
	}
	return referencedFields, nil
}

// AllFields parses attribures to fields
func (e Entity) AllFields() ([]Field, error) {
	var fields []Field
	if err := json.Unmarshal([]byte(e.Fieldsb), &fields); err != nil {
		return nil, errors.Wrapf(err, "error while unmarshalling entity attributes to fields type %q", e.ID)
	}
	return fields, nil
}

func (f Field) isConfig() bool {
	if val, ok := f.Meta["config"]; ok && val == "true" {
		return true
	}
	return false
}

func (f Field) IsReference() bool {
	if f.DataType == TypeReference {
		return true
	}
	return false
}

func (f Field) IsPipe() bool {
	if f.DomType == DomPipeline || f.DomType == DomPlayBook {
		return true
	}
	return false
}

func (f Field) IsNotApplicable() bool {
	if f.DomType == DomNotApplicable {
		return true
	}
	return false
}

func (f Field) DisplayGex() string {
	if val, ok := f.Meta["display_gex"]; ok {
		return val
	}
	return ""
}

func populateBonds(accountID, srcEntityId string, fields []Field) []relationship.Relationship {
	relationships := make([]relationship.Relationship, 0)
	for _, f := range fields {
		if f.IsReference() { // TODO: also check if customer explicitly asks for it. Don't do this for all the reference fields
			relationships = append(relationships, relationship.Relationship{
				RelationshipID: uuid.New().String(),
				AccountID:      accountID,
				SrcEntityID:    srcEntityId,
				DstEntityID:    f.RefID,
				FieldID:        f.Key,
				Type:           relationship.TypeBond,
			})
		}
	}
	return relationships
}

func populateAssociation(accountID, srcEntityId, dstEntityId string) (string, []relationship.Relationship) {
	relationships := make([]relationship.Relationship, 0)
	relationshipID := uuid.New().String()
	relationships = append(relationships, relationship.Relationship{
		RelationshipID: relationshipID,
		AccountID:      accountID,
		SrcEntityID:    srcEntityId,
		DstEntityID:    dstEntityId,
		FieldID:        relationship.FieldAssociationKey,
		Type:           relationship.TypeAssociation,
	}, relationship.Relationship{
		RelationshipID: relationshipID,
		AccountID:      accountID,
		SrcEntityID:    dstEntityId,
		DstEntityID:    srcEntityId,
		FieldID:        relationship.FieldAssociationKey,
		Type:           relationship.TypeAssociation,
	})
	return relationshipID, relationships
}
