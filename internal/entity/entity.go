package entity

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific entity is requested but does not exist.
	ErrEntityNotFound = errors.New("Entity not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidEntityID = errors.New("ID is not in its proper form")
)

// List retrieves a list of existing entities for the team associated from the database.
func List(ctx context.Context, accountID, teamID string, categoryIds []int, db *sqlx.DB) ([]Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.entity.List")
	defer span.End()
	entities := []Entity{}
	if len(categoryIds) == 0 {
		const q = `SELECT * FROM entities where account_id = $1 AND (team_id = $2 OR state = $3 OR shared_team_ids @> $4)`
		if err := db.SelectContext(ctx, &entities, q, accountID, teamID, StateAccountLevel, pq.Array([]string{teamID})); err != nil {
			return nil, errors.Wrap(err, "selecting entities for all category")
		}
	} else {
		const q = `SELECT * FROM entities where account_id = $1 AND category = any($2) AND (team_id = $3 OR state = $4  OR shared_team_ids @> $5)`
		if err := db.SelectContext(ctx, &entities, q, accountID, pq.Array(categoryIds), teamID, StateAccountLevel, pq.Array([]string{teamID})); err != nil {
			return nil, errors.Wrap(err, "selecting entities for category")
		}
	}
	return entities, nil
}

func AccountCoreEntities(ctx context.Context, accountID string, db *sqlx.DB) ([]Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.entity.List")
	defer span.End()
	entities := []Entity{}
	const q = `SELECT * FROM entities where account_id = $1 AND is_core = $2`
	if err := db.SelectContext(ctx, &entities, q, accountID, true); err != nil {
		return nil, errors.Wrap(err, "selecting entities for category")
	}
	return entities, nil
}

func AccountEntities(ctx context.Context, accountID string, categoryIds []int, db *sqlx.DB) ([]Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.entity.List")
	defer span.End()
	entities := []Entity{}
	if len(categoryIds) == 0 {
		const q = `SELECT * FROM entities where account_id = $1`
		if err := db.SelectContext(ctx, &entities, q, accountID); err != nil {
			return nil, errors.Wrap(err, "selecting entities for all category")
		}
	} else {
		const q = `SELECT * FROM entities where account_id = $1 AND category = any($2)`
		if err := db.SelectContext(ctx, &entities, q, accountID, pq.Array(categoryIds)); err != nil {
			return nil, errors.Wrap(err, "selecting entities for category")
		}
	}
	return entities, nil
}

func TeamEntities(ctx context.Context, accountID, teamID string, categoryIds []int, db *sqlx.DB) ([]Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.entity.TeamEntities")
	defer span.End()
	entities := []Entity{}
	if len(categoryIds) == 0 {
		const q = `SELECT * FROM entities where account_id = $1 AND team_id = $2`
		if err := db.SelectContext(ctx, &entities, q, accountID, teamID); err != nil {
			return nil, errors.Wrap(err, "selecting entities for category")
		}
	} else {
		const q = `SELECT * FROM entities where account_id = $1 AND team_id = $2 AND category = any($3)`
		if err := db.SelectContext(ctx, &entities, q, accountID, teamID, pq.Array(categoryIds)); err != nil {
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
		ID:            n.ID,
		AccountID:     n.AccountID,
		TeamID:        n.TeamID,
		Name:          n.Name,
		DisplayName:   n.DisplayName,
		Category:      n.Category,
		State:         n.State,
		Tags:          []string{},
		IsPublic:      n.IsPublic,
		IsCore:        n.IsCore,
		IsShared:      n.IsShared,
		SharedTeamIds: []string{},
		Fieldsb:       string(fieldsBytes),
		CreatedAt:     now.UTC(),
		UpdatedAt:     now.UTC().Unix(),
	}

	const q = `INSERT INTO entities
		(entity_id, account_id, team_id, name, display_name, category, state, tags, is_public, is_core, is_shared, shared_team_ids, fieldsb, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	_, err = db.ExecContext(
		ctx, q,
		e.ID, e.AccountID, e.TeamID, e.Name, e.DisplayName, e.Category, e.State, e.Tags, e.IsPublic, e.IsCore, e.IsShared, e.SharedTeamIds, e.Fieldsb,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return Entity{}, errors.Wrap(err, "inserting entity")
	}

	//TODO: do it in the same transaction.
	//TODO: this relationship should happen only if the user explicitly specifies that.
	//may be, we can give add the boolean in the meta to identify that.
	err = relationship.Bonding(ctx, db, e.AccountID, e.ID, refFields(n.Fields))
	if err != nil {
		return Entity{}, errors.Wrap(err, "making bonds")
	}

	return e, nil
}

// Update replaces a item document in the database.
func Update(ctx context.Context, db *sqlx.DB, sdb *database.SecDB, accountID, entityID string, fieldsB string, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.entity.Update")
	defer span.End()

	const q = `UPDATE entities SET
		"fieldsb" = $3,
		"updated_at" = $4
		WHERE account_id = $1 AND entity_id = $2`
	_, err := db.ExecContext(ctx, q, accountID, entityID,
		fieldsB, now.Unix(),
	)
	if err != nil {
		return err
	}

	updatedFields, err := unmarshalFields(fieldsB)
	if err != nil {
		return err
	}
	sdb.ResetEntity(entityID)

	//TODO: do it in the same transaction.
	return relationship.ReBonding(ctx, db, accountID, entityID, refFields(updatedFields))
}

func UpdateSharedTeam(ctx context.Context, db *sqlx.DB, sdb *database.SecDB, aID, eID string, teamIds pq.StringArray, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.entity.UpdateSharedTeam")
	defer span.End()

	const q = `UPDATE entities SET
		"shared_team_ids" = $3,
		"updated_at" = $4 
		WHERE account_id = $1 AND entity_id = $2`
	_, err := db.ExecContext(ctx, q, aID, eID,
		teamIds, now.Unix(),
	)
	sdb.ResetEntity(eID)

	return err
}

func UpdateMarkers(ctx context.Context, db *sqlx.DB, sdb *database.SecDB, aID, eID string, isPublic, isCore, isShared bool, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.entity.UpdateMarkers")
	defer span.End()

	const q = `UPDATE entities SET
		"is_public" = $3,
		"is_core" = $4,
		"is_shared" = $5,
		"updated_at" = $6 
		WHERE account_id = $1 AND entity_id = $2`
	_, err := db.ExecContext(ctx, q, aID, eID,
		isPublic, isCore, isShared, now.Unix(),
	)

	sdb.ResetEntity(eID)

	return err
}

// Retrieve gets the specified entity from the database.
func Retrieve(ctx context.Context, accountID, entityID string, db *sqlx.DB, sdb *database.SecDB) (Entity, error) {

	ctx, span := trace.StartSpan(ctx, "internal.entity.Retrieve")
	defer span.End()

	var e Entity
	if _, err := uuid.Parse(entityID); err != nil {
		return Entity{}, err
	}

	encodedEntity, err := sdb.RetriveEntity(entityID)
	if err == nil {
		err = json.Unmarshal([]byte(encodedEntity), &e)
		if err == nil {
			return e, err
		}
	}

	const q = `SELECT * FROM entities WHERE account_id = $1 AND entity_id = $2`
	if err := db.GetContext(ctx, &e, q, accountID, entityID); err != nil {
		if err == sql.ErrNoRows {
			return Entity{}, ErrEntityNotFound
		}

		return Entity{}, errors.Wrapf(err, "selecting entity %q", entityID)
	}

	sdb.SetEntity(e.ID, e.Encode())

	return e, nil
}

func RetrieveByName(ctx context.Context, accountID, entityName string, db *sqlx.DB) (Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.entity.RetrieveByName")
	defer span.End()

	var e Entity
	const q = `SELECT * FROM entities WHERE account_id = $1 AND name = $2`
	if err := db.GetContext(ctx, &e, q, accountID, entityName); err != nil {
		if err == sql.ErrNoRows {
			return Entity{}, ErrEntityNotFound
		}

		return Entity{}, errors.Wrapf(err, "selecting entity %q", entityName)
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

func Delete(ctx context.Context, db *sqlx.DB, accountID, entityID string) error {
	ctx, span := trace.StartSpan(ctx, "internal.entity.Delete")
	defer span.End()

	const q = `DELETE FROM entities WHERE account_id = $1 and entity_id = $2`

	if _, err := db.ExecContext(ctx, q, accountID, entityID); err != nil {
		return errors.Wrapf(err, "deleting entity %s", entityID)
	}

	return nil
}

func FetchIDs(entities []Entity) []string {
	ids := make([]string, 0)
	for _, e := range entities {
		ids = append(ids, e.ID)
	}
	return ids
}

func removeIndex(s []interface{}, index int) []interface{} {
	return append(s[:index], s[index+1:]...)
}

func (entity *Entity) Encode() []byte {
	json, err := json.Marshal(entity)
	if err != nil {
		log.Println("***> unexpected/unhandled error in internal.platform.conversation. when marshaling message. error:", err)
	}

	return json
}
