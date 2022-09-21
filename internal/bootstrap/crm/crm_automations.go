package crm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/base"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func addAutomations(ctx context.Context, b *base.Base) error {

	taskEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityTask)
	if err != nil {
		return err
	}

	contactEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityContacts)
	if err != nil {
		return err
	}

	companyEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityCompanies)
	if err != nil {
		return err
	}

	dealEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityDeals)
	if err != nil {
		return err
	}

	cp, err := salesPipeline(ctx, b, dealEntity, taskEntity)
	if err != nil {
		return err
	}
	err = b.AddPipelines(ctx, cp)
	if err != nil {
		return err
	}
	b.SalesPipelineFlowID = cp.FlowID
	fmt.Println("\tCRM:SAMPLES Pipeline And Its Nodes Created")

	//add workflow when company added
	comWF, err := whenCompanyAdded(ctx, b, companyEntity, dealEntity, taskEntity)
	if err != nil {
		return err
	}
	err = b.AddWorkflows(ctx, comWF)
	if err != nil {
		return err
	}

	//add workflow when contact added
	conWF, err := whenContactAdded(ctx, b, contactEntity, dealEntity, taskEntity)
	if err != nil {
		return err
	}
	err = b.AddWorkflows(ctx, conWF)
	if err != nil {
		return err
	}

	dealExceedsWF, err := whenDealAmountExceeds1000(ctx, b, contactEntity, dealEntity)
	if err != nil {
		return err
	}
	err = b.AddWorkflows(ctx, dealExceedsWF)
	if err != nil {
		return err
	}

	fmt.Println("\tCRM:SAMPLES Workflows And Its Nodes Created")

	return nil
}

func whenCompanyAdded(ctx context.Context, b *base.Base, companyEntity, dealEntity, taskEntity entity.Entity) (*base.CoreWorkflow, error) {
	templateFields, templateMeta := dealTemplates(dealEntity, companyEntity, b.SalesPipelineFlowID)
	dealTemplateBasic, err := b.TemplateAdd(ctx, dealEntity.ID, uuid.New().String(), b.UserID, templateFields, templateMeta, nil)
	if err != nil {
		return nil, err
	}
	cf := &base.CoreWorkflow{
		Name:     "When a company is added",
		ActorID:  companyEntity.ID,
		FlowType: flow.FlowTypeEventCreate,
		Nodes: []*base.CoreNode{
			{
				Name:       "Add Base deal",
				ActorID:    dealEntity.ID,
				ActorName:  "Deal",
				TemplateID: dealTemplateBasic.ID,
				Type:       node.Push,
			},
		},
	}
	return cf, nil
}

func whenContactAdded(ctx context.Context, b *base.Base, contactEntity, dealEntity, taskEntity entity.Entity) (*base.CoreWorkflow, error) {
	templateFields, templateMeta := dealTemplates(dealEntity, contactEntity, b.SalesPipelineFlowID)
	dealTemplateBasic, err := b.TemplateAdd(ctx, dealEntity.ID, uuid.New().String(), b.UserID, templateFields, templateMeta, nil)
	if err != nil {
		return nil, err
	}
	cf := &base.CoreWorkflow{
		Name:     "When a contact is added",
		ActorID:  contactEntity.ID,
		FlowType: flow.FlowTypeEventCreate,
		Nodes: []*base.CoreNode{
			{
				Name:       "Add Base deal",
				ActorID:    dealEntity.ID,
				ActorName:  "Deal",
				TemplateID: dealTemplateBasic.ID,
				Type:       node.Modify,
			},
		},
	}
	return cf, nil
}

func whenDealAmountExceeds1000(ctx context.Context, b *base.Base, contactEntity, dealEntity entity.Entity) (*base.CoreWorkflow, error) {
	templateFields, templateMeta := contactTemplates(contactEntity, dealEntity, b.LeadStatusItemNew.ID)
	relatedContactUpdateTemplate, err := b.TemplateAdd(ctx, contactEntity.ID, uuid.New().String(), b.UserID, templateFields, templateMeta, nil)
	if err != nil {
		return nil, err
	}
	cf := &base.CoreWorkflow{
		Name:     "When a deal amount exceeds $1000",
		ActorID:  dealEntity.ID,
		FlowType: flow.FlowTypeEventUpdate,
		Nodes: []*base.CoreNode{
			{
				Name:       "Update related contacts",
				ActorID:    contactEntity.ID,
				ActorName:  "Contact",
				TemplateID: relatedContactUpdateTemplate.ID,
				Type:       node.Modify,
			},
		},
	}
	return cf, nil
}

func salesPipeline(ctx context.Context, b *base.Base, dealEntity, taskEntity entity.Entity) (*base.CoreWorkflow, error) {
	templateFields, templateMeta := taskTemplates("Schedule a call for", taskEntity, dealEntity, true)
	taskTemplateScheduleCall, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, templateFields, templateMeta, nil)
	if err != nil {
		return nil, err
	}

	templateFields, templateMeta = taskTemplates("Prepare the pricing deck", taskEntity, dealEntity, false)
	taskTemplatePreparePricingDeck, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, templateFields, templateMeta, nil)
	if err != nil {
		return nil, err
	}

	templateFields, templateMeta = taskTemplates("Intimate manager about the deal and confirm the deal amount", taskEntity, dealEntity, false)
	taskTemplateIntimateManager, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, templateFields, templateMeta, nil)
	if err != nil {
		return nil, err
	}

	templateFields, templateMeta = taskTemplates("Prepare the invoice and subscription charges for both monthly and annually", taskEntity, dealEntity, false)
	taskTemplatePrepareInvoice, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, templateFields, templateMeta, nil)
	if err != nil {
		return nil, err
	}

	templateFields, templateMeta = taskTemplates("Hand off to finance team", taskEntity, dealEntity, false)
	taskTemplateHandOff, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, templateFields, templateMeta, nil)
	if err != nil {
		return nil, err
	}

	templateFields, templateMeta = taskTemplates("Describe the lost reason in #general conversation", taskEntity, dealEntity, false)
	taskTemplateAddReason, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, templateFields, templateMeta, nil)
	if err != nil {
		return nil, err
	}
	cp := &base.CoreWorkflow{
		Name:    "Sales Pipeline",
		ActorID: dealEntity.ID,
		Exp:     "",
		Nodes: []*base.CoreNode{
			{
				Name:      "Opportunity",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Deals",
				Nodes: []*base.CoreNode{
					{
						Name:       "Schedule a call",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplateScheduleCall.ID,
					},
				},
			},
			{
				Name:      "Interested",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Deals",
				Nodes: []*base.CoreNode{
					{
						Name:       "Prepare pricing deck",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplatePreparePricingDeck.ID,
					},
				},
			},
			{
				Name:      "Qualified",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Deals",
				Nodes: []*base.CoreNode{
					{
						Name:       "Initimate to manager",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplateIntimateManager.ID,
					},
					{
						Name:       "Create invoice ticket",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplatePrepareInvoice.ID,
					},
				},
			},
			{
				Name:      "Won",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Deals",
				Nodes: []*base.CoreNode{
					{
						Name:       "Hand off to finance",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplateHandOff.ID,
					},
				},
			},
			{
				Name:      "Lost",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Deals",
				Nodes: []*base.CoreNode{
					{
						Name:       "Log lost reason",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplateAddReason.ID,
					},
				},
			},
		},
	}
	return cp, nil
}
