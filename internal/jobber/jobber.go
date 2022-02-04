package jobber

import (
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
)

type Jobber interface {
	EventItemCreated(accountID, entityID, itemID string, source map[string]string, db *sqlx.DB, rp *redis.Pool)
	EventItemUpdated(accountID, entityID, itemID string, newFields, oldFields map[string]interface{}, db *sqlx.DB, rp *redis.Pool)
	AddReminder(accountID, entityID, itemID string, when time.Time, rp *redis.Pool) error
	AddDelay(accountID, entityID, itemID string, meta map[string]interface{}, when time.Time, rp *redis.Pool) error
}
