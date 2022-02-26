package engine

import (
	"context"
	"log"
	"reflect"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func (eng *Engine) executeData(ctx context.Context, n node.Node, db *sqlx.DB, rp *redis.Pool) error {
	// value add the fields with the template item provided in the actuals.
	valueAddedFields, err := valueAdd(ctx, db, n.AccountID, n.ActorID, n.ActualsItemID())
	if err != nil {
		return err
	}

	eng.evaluateFieldValues(ctx, db, n.AccountID, valueAddedFields, n.VariablesMap(), n.StageID)
	ni := item.NewItem{
		ID:        uuid.New().String(),
		AccountID: n.AccountID,
		StageID:   &n.StageID,
		EntityID:  n.ActorID,
		Fields:    itemFields(valueAddedFields),
	}

	switch n.Type {
	case node.Push, node.Task, node.Meeting, node.Email:
		log.Printf("internal.rule.engine.executor_data create new node %+v\n", ni)
		it, err := item.Create(ctx, db, ni, time.Now())
		if err != nil {
			return err
		}
		//n.VarStrMap() is equivalent of passing source entity:item in the usual item create
		eng.Job.EventItemCreated(n.AccountID, it.EntityID, it.ID, n.VarStrMap(), db, rp)
	case node.Modify:
		actualItemID := n.ActualsMap()[n.ActorID]
		it, err := item.Retrieve(ctx, n.ActorID, actualItemID, db)
		if err != nil {
			return err
		}
		_, err = item.UpdateFields(ctx, db, n.ActorID, actualItemID, ni.Fields)
		if err != nil {
			return err
		}
		uit, err := item.Retrieve(ctx, n.ActorID, actualItemID, db)
		if err != nil {
			return err
		}
		eng.Job.EventItemUpdated(n.AccountID, it.EntityID, it.ID, uit.Fields(), it.Fields(), db, rp)
	}

	return err
}

func itemFields(fields []entity.Field) map[string]interface{} {
	params := map[string]interface{}{}
	for _, f := range fields {
		params[f.Key] = f.Value
	}
	return params
}

func (eng *Engine) evaluateFieldValues(ctx context.Context, db *sqlx.DB, accountID string, fields []entity.Field, vars map[string]interface{}, stageID string) {
	for i := 0; i < len(fields); i++ {
		var field = &fields[i]

		//associates the item with the stage if executed via the pipeline stage changes from the job
		if field.IsNode() && stageID != node.NoActor {
			field.Value = []interface{}{stageID}
			continue
		}

		switch field.DataType {
		case entity.TypeReference, entity.TypeList:
			if field.Value != nil {
				evalatedVals := make([]interface{}, 0)
				for _, v := range field.Value.([]interface{}) {
					output := eng.RunFieldExpRenderer(ctx, db, accountID, v.(string), vars)
					if output != nil {
						rt := reflect.TypeOf(output)
						switch rt.Kind() {
						case reflect.Slice, reflect.Array:
							evalatedVals = append(evalatedVals, output.([]interface{})...)
						default:
							evalatedVals = append(evalatedVals, output)
						}
					}
				}
				field.Value = evalatedVals
			}
			//old logic. might be useful
			// if _, ok := vars[field.RefID]; ok {
			// 	field.Value = []interface{}{vars[field.RefID]} // TODO: what happens if the vars has more than one item
			// } else if field.Value == nil {
			// 	field.Value = []interface{}{}
			// }
		default:
			if field.Value != nil {
				switch v := field.Value.(type) {
				case string:
					field.Value = eng.RunFieldExpRenderer(ctx, db, accountID, v, vars)
				}

			}
		}

	}
}
