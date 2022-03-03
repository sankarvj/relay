package job

import (
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/rule/engine"
)

// Jab is testing Job
type Jab struct {
}

func (J Jab) EventItemCreated(accountID, userID, entityID, itemID string, source map[string]string, db *sqlx.DB, rp *redis.Pool) {
	log.Println("*> expected error occurred. dead eventItemCreated at jab")
}
func (J Jab) EventItemUpdated(accountID, userID, entityID, itemID string, newFields, oldFields map[string]interface{}, db *sqlx.DB, rp *redis.Pool) {
	log.Println("*> expected error occurred. dead eventItemUpdated at jab")
}

func (J Jab) AddReminder(accountID, userID, entityID, itemID string, when time.Time, rp *redis.Pool) error {
	log.Println("*> expected error occurred. dead AddReminder at jab")
	return nil
}

func (J Jab) AddDelay(accountID, userID, entityID, itemID string, meta map[string]interface{}, when time.Time, rp *redis.Pool) error {
	log.Println("*> expected error occurred. dead AddDelay at jab")
	return nil
}

func NewJabEngine() *engine.Engine {
	return &engine.Engine{
		Job: Jab{},
	}
}
