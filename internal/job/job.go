package job

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
)

//func's in this package should not throw errors. It should handle errors by re-queue/dl-queue

func OnFieldUpdate(entityID, itemID string, oldFields, newFields map[string]interface{}, db *sqlx.DB) {

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
