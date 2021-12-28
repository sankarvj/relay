package base

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/schema"
)

func (b *Base) AddCRMPipelines(ctx context.Context, contactEntityID, webhookEntityID, delayEntityID, delayItemID string) (string, string, error) {
	//add pipelines
	exp := fmt.Sprintf("{{%s.%s}} eq {Vijay} && {{%s.%s}} gt {98}", contactEntityID, schema.SeedFieldFNameKey, contactEntityID, schema.SeedFieldNPSKey)
	p, err := b.FlowAdd(ctx, uuid.New().String(), contactEntityID, "Sales Pipeline", flow.FlowModePipeLine, flow.FlowConditionEntry, exp)
	if err != nil {
		return "", "", err
	}

	dummyID := "00000000-0000-0000-0000-000000000000"

	sno1, err := b.NodeAdd(ctx, uuid.New().String(), p.ID, dummyID, node.Root, "Opportunity", node.Stage, "", map[string]string{}, dummyID, " Opportunity Deals")
	if err != nil {
		return "", "", err
	}

	_, err = b.NodeAdd(ctx, uuid.New().String(), p.ID, dummyID, sno1.ID, "Deal Won", node.Stage, "{Vijay} eq {Vijay}", map[string]string{}, dummyID, "Won Deals")
	if err != nil {
		return "", "", err
	}

	return p.ID, sno1.ID, nil
}

func (b *Base) AddCSMPipeline(ctx context.Context, projectEntityID string, pipeLineName, node1, node2, node3 string) error {
	//add pipelines
	exp := fmt.Sprintf("")
	p, err := b.FlowAdd(ctx, uuid.New().String(), projectEntityID, pipeLineName, flow.FlowModePipeLine, flow.FlowConditionEntry, exp)
	if err != nil {
		return err
	}

	dummyID := "00000000-0000-0000-0000-000000000000"

	sno1, err := b.NodeAdd(ctx, uuid.New().String(), p.ID, dummyID, node.Root, node1, node.Stage, "", map[string]string{}, dummyID, node1)
	if err != nil {
		return err
	}

	//exp = fmt.Sprintf("{{%s.%s}} eq {paid}", projectEntityID, schema.SeedFieldPlanKey)
	sno2, err := b.NodeAdd(ctx, uuid.New().String(), p.ID, dummyID, sno1.ID, node2, node.Stage, exp, map[string]string{}, dummyID, node2)
	if err != nil {
		return err
	}

	_, err = b.NodeAdd(ctx, uuid.New().String(), p.ID, dummyID, sno2.ID, node3, node.Stage, "", map[string]string{}, dummyID, node3)
	if err != nil {
		return err
	}

	return nil
}

func (b *Base) AddSegments(ctx context.Context, entityID string) error {
	e, err := entity.Retrieve(ctx, b.AccountID, entityID, b.DB)
	if err != nil {
		return err
	}

	if e.Name == schema.SeedContactsEntityName {
		err = addSegmentFlow(ctx, entityID, "All Contacts", "", b)
		if err != nil {
			return err
		}
		fields := e.FieldsIgnoreError()
		for _, f := range fields {
			if f.Key == schema.SeedFieldNPSKey {
				exp := fmt.Sprintf("{{%s.%s}} gt {98}", entityID, schema.SeedFieldNPSKey)
				err = addSegmentFlow(ctx, entityID, "High NPS", exp, b)
				if err != nil {
					return err
				}
			}
		}
	} else if e.Name == schema.SeedCompaniesEntityName {
		err = addSegmentFlow(ctx, entityID, "All Companies", "", b)
		if err != nil {
			return err
		}
	} else if e.Name == schema.SeedDealsEntityName {
		err = addSegmentFlow(ctx, entityID, "All Deals", "", b)
		if err != nil {
			return err
		}
	} else if e.Name == schema.SeedProjectsEntityName {
		err = addSegmentFlow(ctx, entityID, "All Projects", "", b)
		if err != nil {
			return err
		}
	}

	return nil
}

func addSegmentFlow(ctx context.Context, entityID, name, exp string, b *Base) error {
	_, err := b.FlowAdd(ctx, uuid.New().String(), entityID, name, flow.FlowModeSegment, flow.FlowConditionNil, exp)
	if err != nil {
		return err
	}
	return nil
}

func (b *Base) AddAssociations(ctx context.Context, conEid, comEid, deEid, tickEid, emailEid string, conID, comID, dealID, ticketID string, emailKey string) error {
	//contact company association
	associationID, err := b.AssociationAdd(ctx, conEid, comEid)
	if err != nil {
		return err
	}
	err = b.ConnectionAdd(ctx, associationID, conID, comID)
	if err != nil {
		return err
	}

	//ticket deal association
	tdaID, err := b.AssociationAdd(ctx, tickEid, deEid)
	if err != nil {
		return err
	}
	err = b.ConnectionAdd(ctx, tdaID, ticketID, dealID)
	if err != nil {
		return err
	}

	//deal email association
	_, err = b.AssociationAdd(ctx, deEid, emailEid)
	if err != nil {
		return err
	}

	//ticket email association
	_, err = b.AssociationAdd(ctx, tickEid, emailEid)
	if err != nil {
		return err
	}

	return nil
}

func (b *Base) AddEmails(ctx context.Context, contactEntityID string, contactEntityKeyEmail, contactEntityKeyNPS string) error {
	emailConfigEntityItem := entity.EmailConfigEntity{
		AccountID: b.AccountID,
		TeamID:    b.TeamID,
		APIKey:    "9c2d8fbbab5c0ca5de49089c1e9777b3-7fba8a4e-b5d71e35",
		Domain:    integration.DomainMailGun,
		Email:     "vijayasankar.jothi@wayplot.com",
		Common:    "false",
		Owner:     []string{schema.SeedUserID1},
	}
	_, err := entity.SaveFixedEntityItem(ctx, b.AccountID, b.TeamID, schema.SeedUserID1, entity.FixedEntityEmailConfig, "Mail Gun Integration", "vijayasankar.jothi@wayplot.com", integration.TypeMailGun, util.ConvertInterfaceToMap(emailConfigEntityItem), b.DB)
	if err != nil {
		return err
	}

	uniqueDiscoveryID := uuid.New().String()
	emailConfigInboxEntityItem := entity.EmailConfigEntity{
		AccountID: b.AccountID,
		TeamID:    b.TeamID,
		APIKey:    uniqueDiscoveryID,
		Domain:    integration.DomainBaseInbox,
		Email:     fmt.Sprintf("support@%s.wayplot.com", uniqueDiscoveryID),
		Common:    "true",
		Owner:     []string{schema.SeedSystemUserID},
	}
	_, err = entity.SaveFixedEntityItem(ctx, b.AccountID, b.TeamID, schema.SeedSystemUserID, entity.FixedEntityEmailConfig, "Base Inbox Integration", uniqueDiscoveryID, integration.TypeBaseInbox, util.ConvertInterfaceToMap(emailConfigInboxEntityItem), b.DB)
	if err != nil {
		return err
	}

	emailEntityItem := entity.EmailEntity{
		From:      []string{},
		To:        []string{fmt.Sprintf("{{%s.%s}}", contactEntityID, contactEntityKeyEmail)},
		Cc:        []string{},
		Bcc:       []string{},
		Contacts:  []string{},
		Companies: []string{},
		Subject:   fmt.Sprintf("This mail is sent you to tell that your NPS scrore is {{%s.%s}}. We are very proud of you!", contactEntityID, contactEntityKeyNPS),
		Body:      fmt.Sprintf("Hello {{%s.%s}}", contactEntityID, contactEntityKeyEmail),
	}

	_, err = entity.SaveFixedEntityItem(ctx, b.AccountID, b.TeamID, schema.SeedUserID1, entity.FixedEntityEmails, "Cult Mail Template", "", "", util.ConvertInterfaceToMap(emailEntityItem), b.DB)
	if err != nil {
		return err
	}
	return nil
}

func (b *Base) AddLayouts(ctx context.Context, name, entityID string) error {
	e, err := entity.Retrieve(ctx, b.AccountID, entityID, b.DB)
	if err != nil {
		return err
	}
	layoutFields := make(map[string]string, 0)
	for _, f := range e.FieldsIgnoreError() {
		if f.Key == "uuid-00-name" { // confusing? because this should happen only via the UI.
			layoutFields["title"] = f.Key
		} else if f.Key == "uuid-00-owner" {
			layoutFields["owner"] = f.Key
		}
	}
	return b.LayoutAdd(ctx, name, entityID, layoutFields)
}
