package jobber

import (
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
)

type Jobber interface {
	Stream(m *stream.Message) error
	AddReminder(accountID, userID, entityID, itemID string, when time.Time, rp *redis.Pool) error
	AddDelay(accountID, userID, entityID, itemID string, meta map[string]interface{}, when time.Time, rp *redis.Pool) error
	AddVisitor(accountID, visitorID, body string, db *sqlx.DB, rp *redis.Pool) error
}
