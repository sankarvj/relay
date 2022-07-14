package csm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/base"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/crm"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func Boot(ctx context.Context, b *base.Base) error {
	b.LoadFixedEntities(ctx)

	conE, comE, _, err := crm.CreateContactCompanyTaskEntity(ctx, b)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT ConComTask Entity Created")

	// add entity - project
	projectEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityProjects, "Projects", entity.CategoryData, entity.StateTeamLevel, false, true, false, ProjectFields(b.StatusEntity.ID, b.OwnerEntity.ID, b.OwnerEntity.Key("email"), conE.ID, comE.ID, b.FlowEntity.ID, b.NodeEntity.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tCSM:BOOT Projects Entity Created")

	// add entity - meetings
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityMeetings, "Meetings", entity.CategoryMeeting, entity.StateTeamLevel, false, true, false, MeetingFields(conE.ID, comE.ID, projectEntity.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Meetings Entity Created")

	return nil

}

func AddSamples(ctx context.Context, b *base.Base) error {
	projectEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityProjects)
	if err != nil {
		return err
	}

	emailsEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityEmails)
	if err != nil {
		return err
	}

	taskEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityTask)
	if err != nil {
		return err
	}

	streamEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityStream)
	if err != nil {
		return err
	}

	err = AddAutomation(ctx, b)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Automations Created")

	err = b.AddSegments(ctx, projectEntity.ID)
	if err != nil {
		return err
	}

	err = AddAssociations(ctx, b, projectEntity, emailsEntity, streamEntity, taskEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Sample Web Of Associations Created Between All The Above Entities")
	return nil
}

func AddAutomation(ctx context.Context, b *base.Base) error {

	projectEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityProjects)
	if err != nil {
		return err
	}

	taskEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityTask)
	if err != nil {
		return err
	}

	taskTemplate1, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Prepare documents", taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate2, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Schedule a meeting", taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate3, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Review the plan with customer", taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate4, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Setup account", taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate5, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Populate sample data", taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate6, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Walkthrough the features", taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate7, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Analyze the metrics", taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate8, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Setup Integrations", taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate9, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Team Traning", taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate10, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Share access", taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate11, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Walk users through the report", taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate12, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Go live", taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate13, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Collect Feedback", taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskTemplate14, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Hand off and mark the project as completed", taskEntity, projectEntity), nil)
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

	taskUpTemplate1, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Send the oppurtuninty the manager", taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskUpTemplate2, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Prepare a pitch", taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskUpTemplate3, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Reachout to customer", taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskUpTemplate4, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Deliver the proposal", taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskUpTemplate5, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Give negotiation", taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskUpTemplate6, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Collect Requirements", taskEntity, projectEntity), nil)
	if err != nil {
		return err
	}
	taskUpTemplate7, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Hand off to finance", taskEntity, projectEntity), nil)
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

	inviteTemplate, err := b.TemplateAdd(ctx, b.InviteEntity.ID, uuid.New().String(), b.UserID, inviteTemplates("Hi, Welcome to the account", b.InviteEntity, projectEntity), nil)
	if err != nil {
		return err
	}

	fmt.Println("\tCSM:SAMPLES inviteTemplate added")

	cf := &base.CoreWorkflow{
		Name:    "When a new project added",
		ActorID: projectEntity.ID,
		Nodes: []*base.CoreNode{
			{
				Name:       "Invite contact as users to the portal",
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

func AddAssociations(ctx context.Context, b *base.Base, proEid, emailEid, streamEID, taskEID entity.Entity) error {

	//project email association
	_, err := b.AssociationAdd(ctx, proEid.ID, emailEid.ID)
	if err != nil {
		return err
	}

	//project task association
	_, err = b.AssociationAdd(ctx, proEid.ID, taskEID.ID)
	if err != nil {
		return err
	}

	//ASSOCIATE STREAMS
	//project stream association
	_, err = b.AssociationAdd(ctx, proEid.ID, streamEID.ID)
	if err != nil {
		return err
	}

	return nil
}
