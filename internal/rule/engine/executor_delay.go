package engine

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func (eng *Engine) executeDelay(ctx context.Context, n node.Node, rulesetResponse map[string]interface{}, db *sqlx.DB, rp *redis.Pool) error {
	entityFields, err := valueAdd(ctx, db, n.AccountID, n.ActorID, n.ActualsItemID())
	if err != nil {
		return err
	}

	var delayEntityItem entity.DelayEntity
	err = entity.ParseFixedEntity(entityFields, &delayEntityItem)
	if err != nil {
		return err
	}
	actualItemID := n.ActualsMap()[n.ActorID]

	//TODO send it to job queue with a delay
	remindBy := time.Now().Add(time.Minute * time.Duration(delayEntityItem.DelayBy))

	meta := rulesetResponse // overload the ruleset response if already exist
	if meta == nil {
		meta = make(map[string]interface{}, 0)
	}
	meta["trigger_entity_id"] = n.Meta.EntityID
	meta["trigger_item_id"] = n.Meta.ItemID
	meta["trigger_flow_type"] = n.Meta.FlowType
	meta["trigger_flow_id"] = n.FlowID
	meta["trigger_node_id"] = n.ID

	err = eng.Job.AddDelay(n.AccountID, n.ActorID, actualItemID, meta, remindBy, rp)
	return err
}
