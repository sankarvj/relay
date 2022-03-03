package jobber

import (
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
)

type Jobber interface {
	EventItemCreated(accountID, userID, entityID, itemID string, source map[string]string, db *sqlx.DB, rp *redis.Pool)
	EventItemUpdated(accountID, userID, entityID, itemID string, newFields, oldFields map[string]interface{}, db *sqlx.DB, rp *redis.Pool)
	AddReminder(accountID, userID, entityID, itemID string, when time.Time, rp *redis.Pool) error
	AddDelay(accountID, userID, entityID, itemID string, meta map[string]interface{}, when time.Time, rp *redis.Pool) error
}
