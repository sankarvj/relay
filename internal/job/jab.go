package job

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/rule/engine"
)

type Jab struct {
}

func (J Jab) AddConnection(ctx context.Context, db *sqlx.DB, accountID string, base map[string]string, entityID, itemID string, newFields, oldFields []entity.Field) {
	log.Println("EventItemCreated Called At Jab")
}

func NewJabEngine() *engine.Engine {
	return &engine.Engine{
		JJ: Jab{},
	}
}
