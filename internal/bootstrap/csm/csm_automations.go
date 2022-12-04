package csm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/base"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func addAutomation(ctx context.Context, b *base.Base) error {
	projectEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityProjects)
	if err != nil {
		return err
	}

	taskEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityTask)
	if err != nil {
		return err
	}

	taskTemplate1, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Prepare documents", "prepare documents about the client org and needs", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate2, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Schedule a meeting", "Ask the client about the convenient time and schedule a meeting", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate3, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Review the plan with customer", "Explain the complete onboarding plan and get the approval from the customer", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate4, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Setup account", "Create a new account and all the clients who have signed up for the demo", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate5, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Populate sample data", "Populate the sample data needed for the smooth onboarding demo", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate6, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Walkthrough the features", "Walkthrough all the features with the use of demo and add if any needed", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate7, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Analyze the metrics", "Use the client analytics to check if the client is using all the features of the product", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate8, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Setup Integrations", "Once the user is comfortable move on to integrations", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate9, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Team Traning", "Ask the customer to send the real data to the product", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate10, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Share access", "Let the customer invite his team members with necessary roles", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate11, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Walk users through the report", "Send the final status report to the client and the manager for reference", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate12, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Go live", "Remove all the demo data populated before going live", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate13, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Collect Feedback", "Collect feedback using the integrated forms in the workbaseONE", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate14, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Hand off and mark the project as completed", "Hand off the project to success/support team", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}

	cp1 := &base.CoreWorkflow{
		Name:    "Basic Onboarding",
		ActorID: projectEntity.ID,
		Exp:     "",
		Nodes: []*base.CoreNode{
			{
				Name:      "Demo Planning & Preparation",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Prepare Projects",
				Nodes: []*base.CoreNode{
					{
						Name:       "Prepare docs",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate1.ID,
					},
					{
						Name:       "Schedule a meeting with client",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate2.ID,
					},
					{
						Name:       "Review the plan with owner",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate3.ID,
					},
				},
			},
			{
				Name:      "Walk Through",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Demo Projects",
				Nodes: []*base.CoreNode{
					{
						Name:       "Set up an account",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate4.ID,
					},
					{
						Name:       "Populate data",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate5.ID,
					},
					{
						Name:       "Walkthrough key features",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate6.ID,
					},
					{
						Name:       "Analyze the metrics",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate7.ID,
					},
					{
						Name:       "Setup integrations",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate8.ID,
					},
				},
			},
			{
				Name:      "Implementation & Verification",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Implementation Projects",
				Nodes: []*base.CoreNode{
					{
						Name:       "Team Training",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate9.ID,
					},
					{
						Name:       "Share access",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate10.ID,
					},
					{
						Name:       "Walk users",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate11.ID,
					},
					{
						Name:       "Go live",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate12.ID,
					},
				},
			},
			{
				Name:      "Final Delivery",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Final Projects",
				Nodes: []*base.CoreNode{
					{
						Name:       "Collect Feedback",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate13.ID,
					},
					{
						Name:       "Hand off to support",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate14.ID,
					},
				},
			},
		},
	}

	taskUpTemplate1, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Send the oppurtuninty the manager", "Once the customer is validated send the details to the manager", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskUpTemplate2, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Prepare a pitch", "Prepare the pitch deck to solve the customer specific use cases", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskUpTemplate3, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Reachout to customer", "Reach out to the customer to gather the real problems and current working model", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskUpTemplate4, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Deliver the proposal", "Analyse the data provided by the customer and send proposal report", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskUpTemplate5, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Give negotiation", "In this step, show the various pricing plans and guide the customer to the right plan", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskUpTemplate6, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Collect Requirements", "Collect his requirements before creating a new account for him", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskUpTemplate7, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Hand off to finance", "Move the customer to the finance team for payment processing", b.StatusItemOpened.ID, taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}

	err = b.AddPipelines(ctx, cp1)
	if err != nil {
		return err
	}

	cp2 := &base.CoreWorkflow{
		Name:    "Upscale Pipeline",
		ActorID: projectEntity.ID,
		Exp:     "",
		Nodes: []*base.CoreNode{
			{
				Name:      "Opportunity",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Opportunity Projects",
				Nodes: []*base.CoreNode{
					{
						Name:       "Send the oppurtuninty the manager",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskUpTemplate1.ID,
					},
					{
						Name:       "Prepare a pitch",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskUpTemplate2.ID,
					},
					{
						Name:       "Reachout to customer",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskUpTemplate3.ID,
					},
				},
			},
			{
				Name:      "Interested",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Interested Projects",
				Nodes: []*base.CoreNode{
					{
						Name:       "Deliver the proposal",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskUpTemplate4.ID,
					},
					{
						Name:       "Give negotiation",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskUpTemplate5.ID,
					},
				},
			},
			{
				Name:      "Won",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Won Project",
				Nodes: []*base.CoreNode{
					{
						Name:       "Collect Requirements",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskUpTemplate6.ID,
					},
					{
						Name:       "Hand off to finance",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskUpTemplate7.ID,
					},
				},
			},
		},
	}

	err = b.AddPipelines(ctx, cp2)
	if err != nil {
		return err
	}

	fmt.Println("\tCSM:SAMPLES Pipeline And Its Nodes Created")

	inviteTemplate, err := b.TemplateAddWithOutMeta(ctx, b.InviteEntity.ID, uuid.New().String(), b.UserID, inviteTemplates("Hi, Welcome to the account", b.InviteEntity, projectEntity), nil)
	if err != nil {
		return err
	}

	fmt.Println("\tCSM:SAMPLES inviteTemplate added")

	cf := &base.CoreWorkflow{
		Name:     "When a new project added",
		ActorID:  projectEntity.ID,
		FlowType: flow.FlowTypeEventCreate,
		Nodes: []*base.CoreNode{
			{
				Name:       "Invite associated contacts of the project to access his record in the portal",
				ActorID:    b.InviteEntity.ID,
				ActorName:  "Projects",
				TemplateID: inviteTemplate.ID,
				Type:       node.Invite,
			},
		},
	}

	err = b.AddWorkflows(ctx, cf)
	if err != nil {
		return err
	}

	fmt.Println("\tCSM:SAMPLES Workflows And Its Nodes Created")

	return nil
}
