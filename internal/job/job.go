package job

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/connection"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
)

//func's in this package should not throw errors. It should handle errors by re-queue/dl-queue

func OnFieldUpdate(account_id, entityID, itemID string, oldFields, newFields map[string]interface{}, db *sqlx.DB) {
	validateExpressions(context.Background(), db, entityID, itemID, oldFields, newFields)
	updateConnection(context.Background(), db, account_id, entityID, itemID, oldFields, newFields)
}

func OnFieldCreate(account_id, entityID, itemID string, newFields map[string]interface{}, db *sqlx.DB) {
	addConnection(context.Background(), db, account_id, entityID, itemID, newFields)
}

func validateExpressions(ctx context.Context, db *sqlx.DB, entityID, itemID string, oldFields, newFields map[string]interface{}) {
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

		log.Println("The flow trigger has been started\n")
		errs := flow.Trigger(context.Background(), db, nil, itemID, dirtyFlows)

		if len(errs) > 0 {
			log.Println("There is an error while triggering flows...", errs)
		}
	}
}

func addConnection(ctx context.Context, db *sqlx.DB, account_id, entityID, itemID string, newFields map[string]interface{}) {
	relationMap := relationMap(ctx, db, account_id, entityID)
	for k, v := range newFields {
		if r, ok := relationMap[k]; ok {
			c := connection.Connection{
				AccountID:      account_id,
				RelationshipID: r.RelationshipID,
				SrcItemID:      itemID,
				DstItemID:      v.([]string),
			}
			_, err := connection.Create(ctx, db, c)
			if err != nil {
				log.Println("error while adding connection", err)
				return
			}
		}
	}
}

func updateConnection(ctx context.Context, db *sqlx.DB, account_id, entityID, itemID string, oldFields, newFields map[string]interface{}) {
	relationMap := relationMap(ctx, db, account_id, entityID)
	dirtyFields := item.Diff(oldFields, newFields)
	for k, v := range dirtyFields {
		if r, ok := relationMap[k]; ok {
			c := connection.Connection{
				AccountID:      account_id,
				RelationshipID: r.RelationshipID,
				SrcItemID:      itemID,
				DstItemID:      util.ConvertSliceTypeRev(v.([]interface{})),
			}
			err := connection.Update(ctx, db, c)
			if err != nil {
				log.Println("error while updating connection", err)
				return
			}
		}
	}
}

func relationMap(ctx context.Context, db *sqlx.DB, account_id, entityID string) map[string]relationship.Relationship {
	relationMap := make(map[string]relationship.Relationship, 0)
	relationships, err := relationship.Relationships(ctx, db, account_id, entityID)
	if err != nil {
		log.Println("There is an error while selecting relationships...", err)
		return relationMap
	}

	for _, r := range relationships {
		relationMap[r.FieldID] = r
	}
	return relationMap
}
