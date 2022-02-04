package relationship

import (
	"context"
	"database/sql"
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
func Bonding(ctx context.Context, db *sqlx.DB, accountID, srcEntityID string, rFields map[string]Relatable) error {
	relationships := populateBonds(accountID, srcEntityID, rFields)
	return bulkCreate(ctx, db, accountID, relationships)
}

// ReBonding updates the implicit relationships between two entities based on the reference fields on the event of entity update
func ReBonding(ctx context.Context, db *sqlx.DB, accountID, srcEntityID string, rFields map[string]Relatable) error {
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

	err = bulkCreate(ctx, db, accountID, newlyAddedRShips)
	if err != nil {
		return err
	}
	err = bulkUpdate(ctx, db, accountID, updatedRShips)
	if err != nil {
		return err
	}
	err = bulkDelete(ctx, db, accountID, deletedRIds)
	if err != nil {
		return err
	}
	return nil
}

// Associate creates the explicit relationships between two entities given by the customer
// This type of associations are always N:N
func Associate(ctx context.Context, db *sqlx.DB, accountID, srcEntityID, dstEntityID string) (string, error) {
	relationshipID, relationships := populateAssociation(accountID, srcEntityID, dstEntityID)
	return relationshipID, bulkCreate(ctx, db, accountID, relationships)
}

//TODO: implement bulk create
func bulkCreate(ctx context.Context, db *sqlx.DB, accountID string, relationships []Relationship) error {
	for _, r := range relationships {
		_, err := Create(ctx, db, r)
		if err != nil {
			err = errors.Wrapf(err, "Association between entities %s and %s failed", r.SrcEntityID, r.DstEntityID)
			return err
		}
	}
	return nil
}

//TODO: implement bulk update
func bulkUpdate(ctx context.Context, db *sqlx.DB, accountID string, relationships []Relationship) error {
	for _, r := range relationships {
		err := Update(ctx, db, r)
		if err != nil {
			return errors.Wrapf(err, "Association update between entities %s and %s failed", r.SrcEntityID, r.DstEntityID)
		}
	}
	return nil
}

func bulkDelete(ctx context.Context, db *sqlx.DB, accountID string, relationshipIDs []string) error {
	for _, rID := range relationshipIDs {
		err := Delete(ctx, db, accountID, rID)
		if err != nil {
			return errors.Wrapf(err, "Association delete for %s failed", rID)
		}
	}
	return nil
}

// Create adds new relationship with respective types.
func Create(ctx context.Context, db *sqlx.DB, r Relationship) (Relationship, error) {
	ctx, span := trace.StartSpan(ctx, "internal.relationship.Create")
	defer span.End()

	const q = `INSERT INTO relationships
		(relationship_id, parent_rel_id, account_id, src_entity_id, dst_entity_id, field_id, type)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := db.ExecContext(
		ctx, q,
		r.RelationshipID, r.ParentRelID, r.AccountID, r.SrcEntityID, r.DstEntityID, r.FieldID, r.Type,
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

func Retrieve(ctx context.Context, accountID, relationshipID string, db *sqlx.DB) (Relationship, error) {
	ctx, span := trace.StartSpan(ctx, "internal.relationship.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(relationshipID); err != nil {
		return Relationship{}, ErrInvalidID
	}

	var r Relationship
	const q = `SELECT * FROM relationships WHERE account_id = $1 AND relationship_id = $2`
	if err := db.GetContext(ctx, &r, q, accountID, relationshipID); err != nil {
		if err == sql.ErrNoRows {
			return Relationship{}, ErrNotFound
		}

		return Relationship{}, errors.Wrapf(err, "selecting relationship %q", relationshipID)
	}

	return r, nil
}

func RetriveAssociation(ctx context.Context, accountID, srcEntityID, dstEntityID string, db *sqlx.DB) (Relationship, error) {
	ctx, span := trace.StartSpan(ctx, "internal.relationship.RetriveAssociation")
	defer span.End()

	var r Relationship
	const q = `SELECT * FROM relationships WHERE account_id = $1 AND src_entity_id = $2 AND dst_entity_id = $3 AND field_id = $4`
	if err := db.GetContext(ctx, &r, q, accountID, srcEntityID, dstEntityID, FieldAssociationKey); err != nil {
		if err == sql.ErrNoRows {
			return Relationship{}, ErrNotFound
		}

		return Relationship{}, errors.Wrapf(err, "selecting explicit relationship using entity ids %q & %q", srcEntityID, dstEntityID)
	}

	return r, nil
}

// List gets the relationships for the destination entity
func List(ctx context.Context, db *sqlx.DB, accountID, teamID, entityID string) ([]Bond, error) {
	ctx, span := trace.StartSpan(ctx, "internal.relationship.List")
	defer span.End()

	var bonds []Bond
	const q = `SELECT r.relationship_id, r.parent_rel_id, e.display_name, e.category, e.entity_id, r.type FROM relationships as r join entities as e on e.entity_id = r.src_entity_id WHERE e.account_id = $1 AND (e.team_id = $2 OR e.state = $3) AND r.dst_entity_id = $4 AND r.type = $5`

	// can't import entity.StateAccountLevel due to cyclic import error hence hardcoded -- 1
	stateAccountLevel := 1
	if err := db.SelectContext(ctx, &bonds, q, accountID, teamID, stateAccountLevel, entityID, RTypeAbsolute); err != nil {
		return nil, errors.Wrap(err, "selecting bonds/relationships for dst entity")
	}

	//trim bonds by reducing relationships with the same entity id
	relatedEntitesMap := make(map[string]Bond, 0)
	for _, b := range bonds {
		relatedEntitesMap[b.EntityID] = b
	}

	trimmedBonds := []Bond{}
	for _, value := range relatedEntitesMap {
		trimmedBonds = append(trimmedBonds, value)
	}

	return trimmedBonds, nil
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

func populateBonds(accountID, srcEntityId string, referenceFields map[string]Relatable) []Relationship {
	relationships := make([]Relationship, 0)
	for fieldKey, relatable := range referenceFields {
		if srcEntityId == "" || relatable.RefID == "" {
			log.Printf("***> unexpected/expected error occurred. src_entity_id (%s) or ref_entity_id (%s) is empty. bonding skipped \n", srcEntityId, relatable.RefID)
			continue
		}

		//Pick parent ID
		var parentRelationID string
		if len(relationships) > 0 {
			parentRelationID = relationships[len(relationships)-1].RelationshipID
		}

		relationshipID := uuid.New().String()
		if relatable.RType == RTypeAbsolute || relatable.RType == RTypeStraight {
			relationships = append(relationships, Relationship{
				RelationshipID: relationshipID,
				ParentRelID:    &parentRelationID,
				AccountID:      accountID,
				SrcEntityID:    srcEntityId,
				DstEntityID:    relatable.RefID,
				FieldID:        fieldKey,
				Type:           relatable.RType,
			})
		}

		if relatable.RType == RTypeAbsolute || relatable.RType == RTypeReverse {
			relationships = append(relationships, Relationship{
				RelationshipID: relationshipID,
				ParentRelID:    &parentRelationID,
				AccountID:      accountID,
				SrcEntityID:    relatable.RefID,
				DstEntityID:    srcEntityId,
				FieldID:        fieldKey,
				Type:           relatable.RType,
			})
		}

	}
	return relationships
}

func updateBonds(accountID, srcEntityId string, existingRelationshipMap map[string]Relationship, referenceFields map[string]Relatable) ([]Relationship, []Relationship, []string) {
	var deletedRelationshipIDs []string
	updatedRelationships := make([]Relationship, 0)

	for _, relationship := range existingRelationshipMap {
		if relatable, ok := referenceFields[relationship.FieldID]; ok {
			delete(referenceFields, relationship.FieldID)
			if relatable.RefID != relationship.DstEntityID {
				relationship.DstEntityID = relatable.RefID
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
		Type:           RTypeAbsolute,
	}, Relationship{
		RelationshipID: relationshipID,
		AccountID:      accountID,
		SrcEntityID:    dstEntityId,
		DstEntityID:    srcEntityId,
		FieldID:        FieldAssociationKey,
		Type:           RTypeAbsolute,
	})
	return relationshipID, relationships
}

func parentIndex(bs []Bond, parentID string) int {
	parentIndex := 0
	for i, b := range bs {
		if b.RelationshipID == parentID {
			parentIndex = i
			break
		}
	}
	return parentIndex
}
