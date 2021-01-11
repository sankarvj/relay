package relationship

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// Relationships
// - Add a new table to maintain the "REVERSE" reference of the reference field. Query that relationship table to find the displayable child entities. The "STRAIGHT" reference of the parent entity is the "REVERSE" reference of the child entity.
// 1. one-to-many
// If a contact has many tasks. Then the task entity would have the contactID as the reference field with (type = one). So, the relationship will be like (src - task) (dest - contact). In the display panel of the contacts we will show the tasks as the child entity and allow to create the task with contact-id prefilled.
// 2. one-to-many
// If a contact has many assignees. Then the contact entity would have the assigneeID as the reference field with (type = one/zero). So, the relationship will be like (src - contact) (dest - assignee). In the display panel of the contacts we will show the assignees as the field property. In the users panel we will show the associated contacts based on "type=one/zero".
// 3. many-to-many
// If a contact has many deals and a deal has many contacts. The deal will have the multiselect contactID as the reference field with (type = two). So, the relationship will be like (src - deal) (dest - contact). In the display panel of the contacts we will show the deals as the child entity and allow to create the deal with contact-id prefilled (REVERSE). In the similar fashion deals are allowed to associate multiple contacts (STRAIGHT).
// 4. special-case
// Though the events is not an regular entity the relationships still holds true for them.
var (
	// ErrNotFound occurs when the relationship is not found
	ErrNotFound = errors.New("Relationship not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// Bonding creates the implicit relationships between two entities based on the reference fields
// This type of associations are always 1:N
func Bonding(ctx context.Context, db *sqlx.DB, accountID, srcEntityID string, rFields map[string]string) error {
	relationships := populateBonds(accountID, srcEntityID, rFields)
	return BulkCreate(ctx, db, relationships)
}

// Associate creates the explicit relationships between two entities given by the customer
// This type of associations are always N:N
func Associate(ctx context.Context, db *sqlx.DB, accountID, srcEntityID, dstEntityID string) (string, error) {
	relationshipID, relationships := populateAssociation(accountID, srcEntityID, dstEntityID)
	return relationshipID, BulkCreate(ctx, db, relationships)
}

//TODO: implement bulk create
func BulkCreate(ctx context.Context, db *sqlx.DB, relationships []Relationship) error {
	for _, r := range relationships {
		_, err := Create(ctx, db, r)
		if err != nil {
			return errors.Wrapf(err, "Association between entities %s and %s failed", r.SrcEntityID, r.DstEntityID)
		}
	}
	return nil
}

// Create add new relationship with respective types.
func Create(ctx context.Context, db *sqlx.DB, r Relationship) (Relationship, error) {
	ctx, span := trace.StartSpan(ctx, "internal.relationship.Create")
	defer span.End()

	const q = `INSERT INTO relationships
		(relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := db.ExecContext(
		ctx, q,
		r.RelationshipID, r.AccountID, r.SrcEntityID, r.DstEntityID, r.FieldID, r.Type,
	)
	if err != nil {
		return Relationship{}, errors.Wrap(err, "inserting relationships")
	}

	return r, nil
}

// List gets the relationships for the destination entity
func List(ctx context.Context, db *sqlx.DB, accountID, entityID string) ([]Bond, error) {
	ctx, span := trace.StartSpan(ctx, "internal.relationship.List")
	defer span.End()

	var bonds []Bond
	const q = `SELECT r.relationship_id, e.display_name, e.category, e.entity_id, r.type FROM relationships as r join entities as e on e.entity_id = r.src_entity_id WHERE r.account_id = $1 AND r.dst_entity_id = $2`

	if err := db.SelectContext(ctx, &bonds, q, accountID, entityID); err != nil {
		return nil, errors.Wrap(err, "selecting bonds/relationships for dst entity")
	}

	return bonds, nil
}

func RelationshipIDs(ctx context.Context, db *sqlx.DB, accountID, entityID string) ([]RelationshipID, error) {
	ctx, span := trace.StartSpan(ctx, "internal.relationship.List")
	defer span.End()

	var relationshipIDs []RelationshipID
	const q = `SELECT relationship_id, field_id FROM relationships WHERE account_id = $1 AND src_entity_id = $2`

	if err := db.SelectContext(ctx, &relationshipIDs, q, accountID, entityID); err != nil {
		return nil, errors.Wrap(err, "selecting relationship id for src entity")
	}

	return relationshipIDs, nil
}

func populateBonds(accountID, srcEntityId string, referenceFields map[string]string) []Relationship {
	relationships := make([]Relationship, 0)
	for fieldKey, refID := range referenceFields {
		relationships = append(relationships, Relationship{
			RelationshipID: uuid.New().String(),
			AccountID:      accountID,
			SrcEntityID:    srcEntityId,
			DstEntityID:    refID,
			FieldID:        fieldKey,
			Type:           TypeBond,
		})

	}
	return relationships
}

func populateAssociation(accountID, srcEntityId, dstEntityId string) (string, []Relationship) {
	relationships := make([]Relationship, 0)
	relationshipID := uuid.New().String()
	relationships = append(relationships, Relationship{
		RelationshipID: relationshipID,
		AccountID:      accountID,
		SrcEntityID:    srcEntityId,
		DstEntityID:    dstEntityId,
		FieldID:        FieldAssociationKey,
		Type:           TypeAssociation,
	}, Relationship{
		RelationshipID: relationshipID,
		AccountID:      accountID,
		SrcEntityID:    dstEntityId,
		DstEntityID:    srcEntityId,
		FieldID:        FieldAssociationKey,
		Type:           TypeAssociation,
	})
	return relationshipID, relationships
}
