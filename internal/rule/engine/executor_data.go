package engine

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func (eng *Engine) executeData(ctx context.Context, db *sqlx.DB, n node.Node) error {
	// value add the fields with the template item provided in the actuals.
	valueAddedFields, err := valueAdd(ctx, db, n.AccountID, n.ActorID, n.ActualsItemID())
	if err != nil {
		return err
	}

	//special handle to update the stage id to the node field.
	for i := 0; i < len(valueAddedFields); i++ {
		if valueAddedFields[i].IsNode() {
			valueAddedFields[i].Value = []interface{}{n.StageID}
		}
	}

	ni := item.NewItem{
		ID:        uuid.New().String(),
		AccountID: n.AccountID,
		EntityID:  n.ActorID,
	}

	eng.evaluateFieldValues(ctx, db, n.AccountID, valueAddedFields, n.VariablesMap())
	ni.Fields = itemFields(valueAddedFields)

	switch n.Type {
	case node.Push:
		it, err := item.Create(ctx, db, ni, time.Now())
		if err != nil {
			return err
		}
		//n.VarStrMap() is equivalent of passing source entity:item in the usual item create
		eng.Job.AddConnection(n.AccountID, n.VarStrMap(), it.EntityID, it.ID, valueAddedFields, nil, db)
	case node.Modify:
		actualItemID := n.ActualsMap()[n.ActorID]
		_, err := item.Retrieve(ctx, n.ActorID, actualItemID, db)
		if err != nil {
			return err
		}
		_, err = item.UpdateFields(ctx, db, n.ActorID, actualItemID, ni.Fields)
		if err != nil {
			return err
		}
		_, err = item.Retrieve(ctx, n.ActorID, actualItemID, db)
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

func (eng *Engine) evaluateFieldValues(ctx context.Context, db *sqlx.DB, accountID string, fields []entity.Field, vars map[string]interface{}) {
	for i := 0; i < len(fields); i++ {
		var field = &fields[i]
		switch fields[i].DataType {
		case entity.TypeString:
			if field.Value != nil {
				field.Value = eng.RunExpRenderer(ctx, db, accountID, field.Value.(string), vars)
			}
		case entity.TypeReference:
			if _, ok := vars[field.RefID]; ok {
				field.Value = []interface{}{vars[field.RefID]} // what happens if the vars has more than one item
			} else if field.Value == nil {
				log.Println("field.Name-----> ", field.Name)
				log.Println("field.Value-----> ", field.Value)
				field.Value = []interface{}{}
			}
		}
	}
}
