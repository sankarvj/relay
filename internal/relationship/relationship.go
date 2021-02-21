package relationship

import (
	"context"
	"log"

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
	return BulkCreate(ctx, db, accountID, relationships)
}

// ReBonding updates the implicit relationships between two entities based on the reference fields on the event of entity update
func ReBonding(ctx context.Context, db *sqlx.DB, accountID, srcEntityID string, rFields map[string]string) error {
	existingRelationships, err := Relationships(ctx, db, accountID, srcEntityID)
	if err != nil {
		return err
	}

	existingRelationshipMap := make(map[string]Relationship, 0)
	for _, relationship := range existingRelationships {
		if relationship.FieldID != FieldAssociationKey {
			existingRelationshipMap[relationship.FieldID] = relationship
		}
	}
	newlyAddedRShips, updatedRShips, deletedRIds := updateBonds(accountID, srcEntityID, existingRelationshipMap, rFields)

	err = BulkCreate(ctx, db, accountID, newlyAddedRShips)
	if err != nil {
		return err
	}
	err = BulkUpdate(ctx, db, accountID, updatedRShips)
	if err != nil {
		return err
	}
	err = BulkDelete(ctx, db, accountID, deletedRIds)
	if err != nil {
		return err
	}
	return nil
}

// Associate creates the explicit relationships between two entities given by the customer
// This type of associations are always N:N
func Associate(ctx context.Context, db *sqlx.DB, accountID, srcEntityID, dstEntityID string) (string, error) {
	relationshipID, relationships := populateAssociation(accountID, srcEntityID, dstEntityID)
	return relationshipID, BulkCreate(ctx, db, accountID, relationships)
}

//TODO: implement bulk create
func BulkCreate(ctx context.Context, db *sqlx.DB, accountID string, relationships []Relationship) error {
	for _, r := range relationships {
		_, err := Create(ctx, db, r)
		if err != nil {
			err = errors.Wrapf(err, "Association between entities %s and %s failed", r.SrcEntityID, r.DstEntityID)
			log.Println(err)
			return err
		}
	}
	return nil
}

//TODO: implement bulk update
func BulkUpdate(ctx context.Context, db *sqlx.DB, accountID string, relationships []Relationship) error {
	for _, r := range relationships {
		err := Update(ctx, db, r)
		if err != nil {
			return errors.Wrapf(err, "Association update between entities %s and %s failed", r.SrcEntityID, r.DstEntityID)
		}
	}
	return nil
}

func BulkDelete(ctx context.Context, db *sqlx.DB, accountID string, relationshipIDs []string) error {
	for _, rID := range relationshipIDs {
		err := Delete(ctx, db, accountID, rID)
		if err != nil {
			return errors.Wrapf(err, "Association delete for %s failed", rID)
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

func Update(ctx context.Context, db *sqlx.DB, r Relationship) error {
	ctx, span := trace.StartSpan(ctx, "internal.relationship.Update")
	defer span.End()

	const q = `UPDATE relationships SET
		"dst_entity_id" = $3,
		WHERE account_id = $1 AND relationship_id = $2`
	_, err := db.ExecContext(ctx, q, r.AccountID, r.RelationshipID, r.DstEntityID)
	if err != nil {
		return err
	}

	return nil
}

func Delete(ctx context.Context, db *sqlx.DB, accountID, relationshipID string) error {
	ctx, span := trace.StartSpan(ctx, "internal.relationship.Delete")
	defer span.End()

	const q = `DELETE FROM relationships WHERE account_id = $1 and relationship_id =$2`

	if _, err := db.ExecContext(ctx, q, accountID, relationshipID); err != nil {
		return errors.Wrapf(err, "deleting relationship %s", relationshipID)
	}

	return nil
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

func Relationships(ctx context.Context, db *sqlx.DB, accountID, entityID string) ([]Relationship, error) {
	ctx, span := trace.StartSpan(ctx, "internal.relationship.List")
	defer span.End()

	var relationships []Relationship
	const q = `SELECT * FROM relationships WHERE account_id = $1 AND src_entity_id = $2`

	if err := db.SelectContext(ctx, &relationships, q, accountID, entityID); err != nil {
		return nil, errors.Wrap(err, "selecting relationships for src entity")
	}

	return relationships, nil
}

func populateBonds(accountID, srcEntityId string, referenceFields map[string]string) []Relationship {
	relationships := make([]Relationship, 0)
	for fieldKey, refID := range referenceFields {
		if srcEntityId == "" || refID == "" {
			log.Printf("either src_entity_id (%s) or ref_entity_id (%s) is empty. Bonding skipped", srcEntityId, refID)
			continue
		}
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

func updateBonds(accountID, srcEntityId string, existingRelationshipMap map[string]Relationship, referenceFields map[string]string) ([]Relationship, []Relationship, []string) {
	var deletedRelationshipIDs []string
	updatedRelationships := make([]Relationship, 0)

	for _, relationship := range existingRelationshipMap {
		if value, ok := referenceFields[relationship.FieldID]; ok {
			delete(referenceFields, relationship.FieldID)
			if value != relationship.DstEntityID {
				relationship.DstEntityID = value
				updatedRelationships = append(updatedRelationships, relationship)
			}
		} else {
			deletedRelationshipIDs = append(deletedRelationshipIDs, relationship.RelationshipID)
		}
	}

	return populateBonds(accountID, srcEntityId, referenceFields), updatedRelationships, deletedRelationshipIDs
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