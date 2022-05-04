package job

import (
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/rule/engine"
)

// Jab is testing Job
type Jab struct {
}

func (J Jab) Stream(m *stream.Message) error {
	return nil
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
