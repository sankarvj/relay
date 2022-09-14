package jobber

import (
	"time"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
)

type Jobber interface {
	Stream(m *stream.Message) error
	AddReminder(accountID, userID, entityID, itemID string, when time.Time, sdb *database.SecDB) error
	AddDelay(accountID, userID, entityID, itemID string, meta map[string]interface{}, when time.Time, sdb *database.SecDB) error
	AddVisitor(accountID, visitorID, body string, db *sqlx.DB, sdb *database.SecDB) error
	AddMember(accountID, memberID, userName, userEmail, body string, db *sqlx.DB, sdb *database.SecDB) error
}
