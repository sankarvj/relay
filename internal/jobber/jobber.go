package jobber

import (
	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
)

type Jobber interface {
	EventItemCreated(accountID, entityID, itemID string, source map[string]string, db *sqlx.DB, rp *redis.Pool)
	EventItemUpdated(accountID, entityID, itemID string, newFields, oldFields map[string]interface{}, db *sqlx.DB, rp *redis.Pool)
}
