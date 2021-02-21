package engine

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func executeDelay(ctx context.Context, db *sqlx.DB, n node.Node) error {
	entityFields, err := mergeActualsWithActor(ctx, db, n.AccountID, n.ActorID, n.ActualsItemID())
	if err != nil {
		return err
	}

	var delayEntityItem entity.DelayEntity
	err = entity.ParseFixedEntity(entityFields, &delayEntityItem)
	if err != nil {
		return err
	}

	//TODO send it to job queue with a delay
	log.Printf("The delay entity %v", delayEntityItem)

	return err
}
