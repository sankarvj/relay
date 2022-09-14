package job

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/rule/engine"
)

// Jab is testing Job
type Jab struct {
}

func (J Jab) Stream(m *stream.Message) error {
	return nil
}

func (J Jab) AddReminder(accountID, userID, entityID, itemID string, when time.Time, sdb *database.SecDB) error {
	log.Println("*> expected error occurred. dead AddReminder at jab")
	return nil
}

func (J Jab) AddDelay(accountID, userID, entityID, itemID string, meta map[string]interface{}, when time.Time, sdb *database.SecDB) error {
	log.Println("*> expected error occurred. dead AddDelay at jab")
	return nil
}

func (J Jab) AddVisitor(accountID, visitorID, body string, db *sqlx.DB, sdb *database.SecDB) error {
	log.Println("*> expected error occurred. dead AddVisitor at jab")
	return nil
}

func (J Jab) AddMember(accountID, memberID, userName, userEmail, body string, db *sqlx.DB, sdb *database.SecDB) error {
	log.Println("*> expected error occurred. dead AddMember at jab")
	return nil
}

func NewJabEngine() *engine.Engine {
	return &engine.Engine{
		Job: Jab{},
	}
}
