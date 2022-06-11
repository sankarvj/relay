package csm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/base"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/schema"
)

func Boot(ctx context.Context, b *base.Base) error {
	b.LoadFixedEntities(ctx)

	// add entity - project
	projectEntity, err := b.EntityAdd(ctx, uuid.New().String(), schema.SeedProjectsEntityName, "Projects", entity.CategoryData, entity.StateTeamLevel, ProjectFields(b.StatusEntity.ID, b.OwnerEntity.ID, b.OwnerEntity.Key("email"), b.ContactEntity.ID, b.CompanyEntity.ID, b.FlowEntity.ID, b.NodeEntity.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tCSM:BOOT Deals Entity Created")

	// add entity - task
	_, err = b.EntityAdd(ctx, uuid.New().String(), schema.SeedTasksEntityName, "Tasks", entity.CategoryTask, entity.StateTeamLevel, TaskFields(b.ContactEntity.ID, b.CompanyEntity.ID, projectEntity.ID, b.StatusEntity.ID, b.NodeEntity.ID, b.StatusItemOpened.ID, b.StatusItemClosed.ID, b.StatusItemOverDue.ID, b.TypeEntity.ID, b.TypeItemEmail.ID, b.TypeItemTodo.ID, b.EmailsEntity.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tCSM:BOOT Tasks Entity Created")

	// add entity - meetings
	_, err = b.EntityAdd(ctx, uuid.New().String(), schema.SeedMeetingsEntityName, "Meetings", entity.CategoryMeeting, entity.StateTeamLevel, MeetingFields(b.ContactEntity.ID, b.CompanyEntity.ID, projectEntity.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Meetings Entity Created")

	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityStream, "Streams", entity.CategoryStream, entity.StateTeamLevel, forms.StreamFields())
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Streams Entity Created")

	return nil

}

func AddSamples(ctx context.Context, b *base.Base) error {
	projectEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, schema.SeedProjectsEntityName)
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
	return nil
}

func AddAutomation(ctx context.Context, b *base.Base) error {

	projectEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, schema.SeedProjectsEntityName)
	if err != nil {
		return err
	}

	taskEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, schema.SeedTasksEntityName)
	if err != nil {
		return err
	}

	taskTemplate, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates(projectEntity), nil)
	if err != nil {
		return err
	}

	cp1 := &base.CoreWorkflow{
		Name:    "Basic Onboarding",
		ActorID: projectEntity.ID,
		Exp:     "",
		Nodes: []*base.CoreNode{
			{
				Name:      "Opportunity",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Opportunity Deals",
				Nodes: []*base.CoreNode{
					{
						Name:       "Send a e-mail",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate.ID,
					},
				},
			},
			{
				Name:      "Interested",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Opportunity Deals",
				Nodes: []*base.CoreNode{
					{
						Name:       "Schedule a call",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate.ID,
					},
				},
			},
			{
				Name:      "Qualified",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Opportunity Deals",
				Nodes: []*base.CoreNode{
					{
						Name:       "Prepare for a demo",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate.ID,
					},
				},
			},
			{
				Name:      "Won",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Won Deals",
				Nodes: []*base.CoreNode{
					{
						Name:       "Hand off to finance",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate.ID,
					},
				},
			},
		},
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
				ActorName: "Opportunity Deals",
				Nodes: []*base.CoreNode{
					{
						Name:       "Send a e-mail",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate.ID,
					},
				},
			},
			{
				Name:      "Interested",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Opportunity Deals",
				Nodes: []*base.CoreNode{
					{
						Name:       "Schedule a call",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate.ID,
					},
				},
			},
			{
				Name:      "Qualified",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Opportunity Deals",
				Nodes: []*base.CoreNode{
					{
						Name:       "Prepare for a demo",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate.ID,
					},
				},
			},
			{
				Name:      "Won",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "Won Deals",
				Nodes: []*base.CoreNode{
					{
						Name:       "Hand off to finance",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate.ID,
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
	return nil
}
