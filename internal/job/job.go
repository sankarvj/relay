package job

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/connection"
	"gitlab.com/vjsideprojects/relay/internal/email"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
)

//func's in this package should not throw errors. It should handle errors by re-queue/dl-queue

func EventItemUpdated(accountID, entityID, itemID string, oldFields, newFields map[string]interface{}, db *sqlx.DB) {
	ctx := context.Background()
	validateWorkflows(ctx, db, entityID, itemID, oldFields, newFields)
	updateConnection(ctx, db, accountID, entityID, itemID, oldFields, newFields)
}

func EventItemCreated(accountID, entityID, itemID string, newFields map[string]interface{}, db *sqlx.DB) {
	ctx := context.Background()

	e, err := entity.Retrieve(ctx, accountID, entityID, db)
	if err != nil {
		log.Println("error while retriving entity on job", err)
		return
	}
	valueAddedFields := entity.ValueAddFields(e.FieldsIgnoreError(), newFields)
	//validateWorkflows(ctx, db, entityID, itemID, oldFields, newFields)
	addConnection(ctx, db, accountID, entityID, itemID, valueAddedFields)

	switch e.Category {
	case entity.CategoryEmail:

		reference.UpdateChoicesWrapper(ctx, db, accountID, valueAddedFields)
		err = email.SendMail(ctx, accountID, e.ID, itemID, valueAddedFields, db)
	}
	if err != nil {
		log.Println("error while performing the job", err)
	}

}

func validateWorkflows(ctx context.Context, db *sqlx.DB, entityID, itemID string, oldFields, newFields map[string]interface{}) {
	// log.Println("entityID...", entityID)
	// log.Println("itemID...", itemID)
	// log.Println("oldFields...", oldFields)
	// log.Println("newFields...", newFields)
	flows, err := flow.List(context.Background(), []string{entityID}, -1, db)
	if err != nil {
		log.Println("There is an error while selecting flows...", err)
	}
	dirtyFlows := flow.DirtyFlows(context.Background(), flows, oldFields, newFields)

	log.Printf("This update triggers %d flows", len(dirtyFlows))
	if len(dirtyFlows) > 0 {
		log.Print("Tick...\nTick...\nTick...\nTick...\nTick...\nTick...\n")

		log.Println("The flow trigger has been started")
		errs := flow.Trigger(context.Background(), db, nil, itemID, dirtyFlows)

		if len(errs) > 0 {
			log.Println("There is an error while triggering flows...", errs)
		}
	}
}

func addConnection(ctx context.Context, db *sqlx.DB, accountID, entityID, itemID string, valueAddedFields []entity.Field) {
	relationMap := relationMap(ctx, db, accountID, entityID)
	for _, field := range valueAddedFields {

		//skip for node & flow
		if field.IsFlow() || field.IsNode() {
			continue
		}

		if relationshipID, ok := relationMap[field.Key]; ok {
			if field.Value == nil || field.Value == "" || len(field.Value.([]interface{})) == 0 {
				continue
			}
			//TODO: use batch create
			for _, dstItemID := range field.Value.([]interface{}) {
				c := connection.Connection{
					AccountID:      accountID,
					RelationshipID: relationshipID,
					SrcItemID:      itemID,
					DstItemID:      dstItemID.(string),
				}

				_, err := connection.Create(ctx, db, c)
				if err != nil {
					log.Println("error while adding connection", err)
					return
				}
			}
		}
	}
}

func updateConnection(ctx context.Context, db *sqlx.DB, account_id, entityID, itemID string, oldFields, newFields map[string]interface{}) {
	relationMap := relationMap(ctx, db, account_id, entityID)
	dirtyFields := item.Diff(oldFields, newFields)
	for k, v := range dirtyFields {
		if relationshipID, ok := relationMap[k]; ok {
			oldDstItemIds := oldFields[k].([]interface{})
			newDstItemIds := v.([]interface{})

			deletedItems, newItems := item.CompareItems(oldDstItemIds, newDstItemIds)
			if len(deletedItems) > 0 {
				//TODO: use batch delete
				for _, deletedItem := range deletedItems {
					err := connection.Delete(ctx, db, relationshipID, deletedItem.(string))
					if err != nil {
						log.Println("error while deleting connection", err)
						return
					}
				}
			}

			if len(newItems) > 0 {
				//TODO: use batch create
				for _, dstItemID := range newItems {
					c := connection.Connection{
						AccountID:      account_id,
						RelationshipID: relationshipID,
						SrcItemID:      itemID,
						DstItemID:      dstItemID.(string),
					}
					_, err := connection.Create(ctx, db, c)
					if err != nil {
						log.Println("error while adding connection", err)
						return
					}
				}
			}

		}
	}
}

func relationMap(ctx context.Context, db *sqlx.DB, accountID, entityID string) map[string]string {
	relationMap := make(map[string]string, 0)
	relationships, err := relationship.Relationships(ctx, db, accountID, entityID)
	if err != nil {
		log.Println("There is an error while selecting relationships...", err)
		return relationMap
	}

	for _, r := range relationships {
		relationMap[r.FieldID] = r.RelationshipID
	}
	return relationMap
}
