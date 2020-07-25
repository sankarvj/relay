package engine

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func executeDelay(ctx context.Context, db *sqlx.DB, n node.Node) error {
	entityFields, err := fields(ctx, db, n.ActorID)
	if err != nil {
		return err
	}
	entityFields, err = fillItemFieldValues(ctx, db, entityFields, n.ActualsMap()[n.ActorID])
	if err != nil {
		return err
	}

	delayEntity, err := entity.ParseDelayEntity(namedFieldsMap(entityFields))
	if err != nil {
		return err
	}

	//TODO send it to job queue with a delay
	log.Printf("The delay entity %v", delayEntity)

	return err
}
