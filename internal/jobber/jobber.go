package jobber

import (
	"context"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

type Jobber interface {
	AddConnection(ctx context.Context, db *sqlx.DB, accountID string, base map[string]string, entityID, itemID string, newFields, oldFields []entity.Field)
}
