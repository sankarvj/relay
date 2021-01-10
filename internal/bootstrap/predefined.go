package bootstrap

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific entity is requested but does not exist.
	ErrNotFound      = errors.New("Entity not found")
	TokenKeyNotFound = errors.New("Token key not found in user entity")
)

type PreDefined struct {
	itemID   string
	entityID string
	fields   []entity.Field
}

func RetrievePreDefinedEntity(ctx context.Context, db *sqlx.DB, accountID string, systemEntityName string) (entity.Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.predefined.RetrieveUserEntity")
	defer span.End()

	var e entity.Entity
	const q = `SELECT * FROM entities WHERE account_id = $1 AND name = $2 LIMIT 1`
	if err := db.GetContext(ctx, &e, q, accountID, systemEntityName); err != nil {
		if err == sql.ErrNoRows {
			return entity.Entity{}, ErrNotFound
		}

		return entity.Entity{}, errors.Wrapf(err, "selecting system entity %q", systemEntityName)
	}

	return e, nil
}

func CurrentUserItem(ctx context.Context, accountID string, db *sqlx.DB) (entity.UserEntity, *PreDefined, error) {
	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return entity.UserEntity{}, nil, err
	}
	ownerEntity, err := RetrievePreDefinedEntity(ctx, db, accountID, OwnerEntity)
	if err != nil {
		return entity.UserEntity{}, nil, err
	}
	it, err := item.Retrieve(ctx, ownerEntity.ID, currentUserID, db)
	if err != nil {
		return entity.UserEntity{}, nil, err
	}

	entityFields, err := ownerEntity.Fields()
	if err != nil {
		return entity.UserEntity{}, nil, err
	}

	entityFields = entity.FillFieldValues(entityFields, it.Fields())
	userEntity, err := entity.ParseUserEntity(entity.NamedFieldsMap(entityFields))

	predefined := &PreDefined{
		itemID:   it.ID,
		entityID: it.EntityID,
		fields:   entityFields,
	}
	return userEntity, predefined, err
}

func updateFields(ctx context.Context, accountID, entityID, itemID string, fields []entity.Field, namedItem interface{}, db *sqlx.DB) func() error {
	return func() error {
		return item.UpdateFields(ctx, db, entityID, itemID, entity.KeyedFieldsMap(fields, util.ConvertInterfaceToMap(namedItem)))
	}
}

func SaveToken(ctx context.Context, accountID, token string, db *sqlx.DB) error {
	cu, pd, err := CurrentUserItem(ctx, accountID, db)
	if err != nil {
		return err
	}
	cu.Gtoken = token
	return item.UpdateFields(ctx, db, pd.entityID, pd.itemID, entity.KeyedFieldsMap(pd.fields, util.ConvertInterfaceToMap(cu)))
}
