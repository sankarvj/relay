package engine

import (
	"context"
	"log"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/connection"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

func (eng *Engine) executeData(ctx context.Context, n node.Node, db *sqlx.DB, sdb *database.SecDB) error {
	log.Printf("internal.rule.engine.executeData %s - %s \n", n.ActorID, n.ActualsItemID())
	// value add the fields with the template item provided in the actuals.
	valueAddedFields, err := valueAdd(ctx, db, sdb, n.AccountID, n.ActorID, n.ActualsItemID())
	if err != nil {
		return err
	}

	eng.evaluateFieldValues(ctx, db, sdb, n.AccountID, valueAddedFields, n.VariablesMap(), n.StageID)

	templateItem := item.NewItem{
		ID:        uuid.New().String(),
		UserID:    util.String(user.UUID_ENGINE_USER),
		AccountID: n.AccountID,
		StageID:   &n.StageID,
		EntityID:  n.ActorID,
		Fields:    itemFields(valueAddedFields),
	}

	switch n.Type {
	case node.Push, node.Task, node.Meeting, node.Email:
		it, err := item.Create(ctx, db, templateItem, time.Now())
		if err != nil {
			return err
		}
		eng.Job.Stream(stream.NewCreteItemMessage(ctx, db, n.AccountID, user.UUID_ENGINE_USER, it.EntityID, it.ID, n.VarStrMap()))
		//n.VarStrMap() is equivalent of passing source entity:item in the usual item create
	case node.Modify:
		actualItemID := n.ActualsItemID()

		// update the trigger entity/item if the actor/trigger are same
		// when a deal is updated change the status of the deal
		if n.ActorID == n.Meta.EntityID {
			err = eng.updateItemFields(ctx, n.AccountID, n.ActorID, actualItemID, templateItem, db, sdb)
		} else { // when a deal is updated, make changes it all of its related contact.
			err = eng.updateRelatedItems(ctx, n.AccountID, n.Meta.EntityID, n.Meta.ItemID, n.ActorID, templateItem, "", db, sdb)
		}

	}

	return err
}

func (eng *Engine) updateItemFields(ctx context.Context, accountID, actorEntityID, actorItemID string, templateItem item.NewItem, db *sqlx.DB, sdb *database.SecDB) error {

	e, err := entity.Retrieve(ctx, accountID, actorEntityID, db, sdb)
	if err != nil {
		log.Println("***>***> EventItemUpdated: unexpected/unhandled error occurred when retriving entity inside job. error:", err)
		return err
	}

	it, err := item.Retrieve(ctx, actorEntityID, actorItemID, db)
	if err != nil {
		return err
	}

	fields := e.FieldsIgnoreError()
	itemFields := it.Fields()

	for _, f := range fields {
		if templateItem.Fields[f.Key] != "" && templateItem.Fields[f.Key] != nil {
			itemFields[f.Key] = f.CalcFunc().Calc(itemFields[f.Key], templateItem.Fields[f.Key])
		}
	}

	_, err = item.UpdateFields(ctx, db, actorEntityID, actorItemID, itemFields)
	if err != nil {
		return err
	}
	uit, err := item.Retrieve(ctx, actorEntityID, actorItemID, db)
	if err != nil {
		return err
	}

	eng.Job.Stream(stream.NewUpdateItemMessage(ctx, db, accountID, user.UUID_ENGINE_USER, it.EntityID, it.ID, uit.Fields(), it.Fields()))
	log.Println("internal.rule.engine.executeData: Item fields updated successfully")
	return nil
}

func (eng *Engine) updateRelatedItems(ctx context.Context, accountID, srcEntityID, srcItemID, actorEntityID string, templateItem item.NewItem, next string, db *sqlx.DB, sdb *database.SecDB) error {
	var err error
	connectedItems, err := connection.AllChild(ctx, db, accountID, srcEntityID, srcItemID, actorEntityID, next)
	if err != nil {
		return err
	}

	for _, childItem := range connectedItems { //each contact needs to get updated with the template.
		err = eng.updateItemFields(ctx, accountID, actorEntityID, childItem.DstItemID, templateItem, db, sdb)
		if err != nil {
			return err
		}
	}

	if len(connectedItems) == 1000 {
		next = connectedItems[len(connectedItems)-1].ConnectionID
		err = eng.updateRelatedItems(ctx, accountID, srcEntityID, srcItemID, actorEntityID, templateItem, next, db, sdb)
		if err != nil {
			return err
		}
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

func (eng *Engine) evaluateFieldValues(ctx context.Context, db *sqlx.DB, sdb *database.SecDB, accountID string, fields []entity.Field, vars map[string]interface{}, stageID string) {
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
					output := eng.RunFieldExpRenderer(ctx, db, sdb, accountID, v.(string), vars)
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
					field.Value = eng.RunFieldExpRenderer(ctx, db, sdb, accountID, v, vars)
				}

			}
		}

	}
}
