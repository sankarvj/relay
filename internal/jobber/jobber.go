package jobber

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

type Jobber interface {
	AddConnection(accountID string, base map[string]string, entityID, itemID string, newFields, oldFields []entity.Field, db *sqlx.DB)
}
