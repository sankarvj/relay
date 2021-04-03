package bootstrap

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func addPipelines(ctx context.Context, db *sqlx.DB, accountID, contactEntityID, mailEntityID, webhookEntityID, delayEntityID, mailItemID, delayItemID string) (string, string, error) {
	//add pipelines
	p, err := FlowAdd(ctx, db, accountID, uuid.New().String(), contactEntityID, "Sales Pipeline", flow.FlowModePipeLine, flow.FlowConditionEntry)
	if err != nil {
		return "", "", err
	}

	dummyID := "00000000-0000-0000-0000-000000000000"

	sno1, err := NodeAdd(ctx, db, accountID, uuid.New().String(), p.ID, dummyID, node.Root, "Opportunity", node.Stage, "", map[string]string{}, dummyID, " Opportunity Deals")
	if err != nil {
		return "", "", err
	}

	_, err = NodeAdd(ctx, db, accountID, uuid.New().String(), p.ID, dummyID, sno1.ID, "Deal Won", node.Stage, "{Vijay} eq {Vijay}", map[string]string{}, dummyID, "Won Deals")
	if err != nil {
		return "", "", err
	}

	// no1, err := NodeAdd(ctx, db, accountID, uuid.New().String(), p.ID, mailEntityID, sno1.ID, "Email", node.Email, "", map[string]string{mailEntityID: mailItemID}, sno1.ID, "Send mail to customer")
	// if err != nil {
	// 	return "", "", err
	// }

	// _, err = NodeAdd(ctx, db, accountID, uuid.New().String(), p.ID, webhookEntityID, no1.ID, "Hook", node.Hook, "", map[string]string{}, sno1.ID, " Hit customer API")
	// if err != nil {
	// 	return "", "", err
	// }

	// _, err = NodeAdd(ctx, db, accountID, uuid.New().String(), p.ID, delayEntityID, sno2.ID, "Delay", node.Delay, "", map[string]string{delayEntityID: delayItemID}, sno2.ID, "Wait for 5 mins")
	// if err != nil {
	// 	return "", "", err
	// }
	return p.ID, sno1.ID, nil
}
