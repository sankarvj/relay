package job

import (
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/rule/engine"
)

// Jab is testing Job
type Jab struct {
}

func (J Jab) AddConnection(accountID string, base map[string]string, entityID, itemID string, newFields, oldFields []entity.Field, db *sqlx.DB) {
	log.Println("DeadAddConnection At Jab")
}

func NewJabEngine() *engine.Engine {
	return &engine.Engine{
		Job: Jab{},
	}
}
