package bootstrap

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/schema"
)

func addPipelines(ctx context.Context, db *sqlx.DB, accountID, contactEntityID, webhookEntityID, delayEntityID, delayItemID string) (string, string, error) {
	//add pipelines
	exp := fmt.Sprintf("{{%s.%s}} eq {Vijay} && {{%s.%s}} gt {98}", contactEntityID, schema.SeedFieldFNameKey, contactEntityID, schema.SeedFieldNPSKey)
	p, err := FlowAdd(ctx, db, accountID, uuid.New().String(), contactEntityID, "Sales Pipeline", flow.FlowModePipeLine, flow.FlowConditionEntry, exp)
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

func addSegments(ctx context.Context, db *sqlx.DB, accountID, entityID string) error {
	e, err := entity.Retrieve(ctx, accountID, entityID, db)
	if err != nil {
		return err
	}

	if e.Name == schema.SeedContactsEntityName {
		err = BootstrapSegments(ctx, db, accountID, entityID, "All Contacts", "")
		if err != nil {
			return err
		}
		fields := e.FieldsIgnoreError()
		for _, f := range fields {
			if f.Key == schema.SeedFieldNPSKey {
				exp := fmt.Sprintf("{{%s.%s}} gt {98}", entityID, schema.SeedFieldNPSKey)
				err = BootstrapSegments(ctx, db, accountID, entityID, "High NPS", exp)
				if err != nil {
					return err
				}
			}
		}
	} else if e.Name == schema.SeedCompaniesEntityName {
		err = BootstrapSegments(ctx, db, accountID, entityID, "All Companies", "")
		if err != nil {
			return err
		}
	} else if e.Name == schema.SeedDealsEntityName {
		err = BootstrapSegments(ctx, db, accountID, entityID, "All Deals", "")
		if err != nil {
			return err
		}
	}

	return nil
}
