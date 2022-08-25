package base

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/schema"
)

func (b *Base) AddPipelines(ctx context.Context, cwf *CoreWorkflow) error {
	stageID := "00000000-0000-0000-0000-000000000000"

	f, err := b.FlowAdd(ctx, uuid.New().String(), cwf.ActorID, cwf.Name, flow.FlowModePipeLine, flow.FlowConditionEntry, cwf.Exp, flow.FlowTypeEventCreate)
	if err != nil {
		return err
	}
	cwf.FlowID = f.ID

	for i := 0; i < len(cwf.Nodes); i++ {
		cn := cwf.Nodes[i]

		parentNodeID := node.Root
		if i > 0 {
			pn := cwf.Nodes[i-1]
			parentNodeID = pn.NodeID
		}

		n, err := b.NodeAdd(ctx, uuid.New().String(), cwf.FlowID, cn.ActorID, parentNodeID, cn.Name, node.Stage, cn.Exp, map[string]string{}, cn.Tokens, stageID, cn.ActorName)
		if err != nil {
			return err
		}
		cn.NodeID = n.ID

		for j := 0; j < len(cn.Nodes); j++ {
			nus := cn.Nodes[j] // node under the stage
			parentNodeID := cn.NodeID
			if j > 0 {
				pn := cn.Nodes[j-1]
				parentNodeID = pn.NodeID
			}

			nusn, err := b.NodeAdd(ctx, uuid.New().String(), cwf.FlowID, nus.ActorID, parentNodeID, nus.Name, node.Push, nus.Exp, map[string]string{nus.ActorID: nus.TemplateID}, nus.Tokens, cn.NodeID, nus.ActorName)
			if err != nil {
				return err
			}
			nus.NodeID = nusn.ID
		}
	}

	return nil
}

func (b *Base) AddWorkflows(ctx context.Context, cwf *CoreWorkflow) error {
	stageID := "00000000-0000-0000-0000-000000000000"
	f, err := b.FlowAdd(ctx, uuid.New().String(), cwf.ActorID, cwf.Name, flow.FlowModeWorkFlow, flow.FlowConditionEntry, "", flow.FlowTypeEventCreate)
	if err != nil {
		return err
	}
	cwf.FlowID = f.ID

	for i := 0; i < len(cwf.Nodes); i++ {
		cn := cwf.Nodes[i]

		parentNodeID := node.Root
		if i > 0 {
			pn := cwf.Nodes[i-1]
			parentNodeID = pn.NodeID
		}

		n, err := b.NodeAdd(ctx, uuid.New().String(), cwf.FlowID, cn.ActorID, parentNodeID, cn.Name, cn.Type, "", map[string]string{cn.ActorID: cn.TemplateID}, cn.Tokens, stageID, cn.ActorName)
		if err != nil {
			return err
		}
		cn.NodeID = n.ID
	}

	return nil
}

func (b *Base) AddSegments(ctx context.Context, entityID string) error {
	e, err := entity.Retrieve(ctx, b.AccountID, entityID, b.DB)
	if err != nil {
		return err
	}

	if e.Name == entity.FixedEntityContacts {
		err = addSegmentFlow(ctx, entityID, "All Contacts", "", b)
		if err != nil {
			return err
		}
		fields := e.FieldsIgnoreError()
		for _, f := range fields {
			if f.Name == "nps_score" {
				exp := fmt.Sprintf("{{%s.%s}} gt {98}", entityID, e.Key("nps_score"))
				err = addSegmentFlow(ctx, entityID, "High NPS", exp, b)
				if err != nil {
					return err
				}
			}
		}
	} else if e.Name == entity.FixedEntityCompanies {
		err = addSegmentFlow(ctx, entityID, "All Companies", "", b)
		if err != nil {
			return err
		}
	} else if e.Name == entity.FixedEntityDeals {
		err = addSegmentFlow(ctx, entityID, "All Deals", "", b)
		if err != nil {
			return err
		}
	} else if e.Name == entity.FixedEntityProjects {
		err = addSegmentFlow(ctx, entityID, "All Projects", "", b)
		if err != nil {
			return err
		}
	}

	return nil
}

func addSegmentFlow(ctx context.Context, entityID, name, exp string, b *Base) error {
	_, err := b.FlowAdd(ctx, uuid.New().String(), entityID, name, flow.FlowModeSegment, flow.FlowConditionNil, exp, flow.FlowTypeUnknown)
	if err != nil {
		return err
	}
	return nil
}

func (b *Base) AddConnections(ctx context.Context, associationID1, associationID2 string, conEid, comEid, deEid, tickEid entity.Entity, conID, comID, dealID, ticketID item.Item) error {
	//contact company association
	// err := b.ConnectionAdd(ctx, associationID1, "Contact", conEid.ID, comEid.ID, conID.ID, comID.ID, comEid.ValueAdd(comID.Fields()), "Created")
	// if err != nil {
	// 	return err
	// }

	// err = b.ConnectionAdd(ctx, associationID2, "Ticket", deEid.ID, tickEid.ID, ticketID.ID, dealID.ID, tickEid.ValueAdd(ticketID.Fields()), "Created")
	// if err != nil {
	// 	return err
	// }

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

	emailEntityItem := entity.EmailEntity{
		From:    []string{},
		To:      []string{fmt.Sprintf("{{%s.%s}}", contactEntityID, contactEntityKeyEmail)},
		Cc:      []string{},
		Bcc:     []string{},
		Subject: fmt.Sprintf("This mail is sent you to tell that your NPS scrore is {{%s.%s}}. We are very proud of you!", contactEntityID, contactEntityKeyNPS),
		Body:    fmt.Sprintf("Hello {{%s.%s}}", contactEntityID, contactEntityKeyEmail),
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
		if f.Name == "first_name" { // confusing? because this should happen only via the UI.
			layoutFields["title"] = f.Key
		} else if f.Name == "owner" {
			layoutFields["owner"] = f.Key
		}
	}
	return b.LayoutAdd(ctx, name, entityID, layoutFields)
}
