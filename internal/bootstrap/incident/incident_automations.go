package incident

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/base"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func addAutomation(ctx context.Context, b *base.Base) error {

	incidentEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityIncidents)
	if err != nil {
		return err
	}

	taskEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityTask)
	if err != nil {
		return err
	}

	taskTemplate1, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Prepare documents", "Document the cause of fix of the incident", b.StatusItemOpened.ID, taskEntity), nil)
	if err != nil {
		return err
	}

	taskTemplate2, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("Notify managers", "Notify the progress of incidents to managers", b.StatusItemOpened.ID, taskEntity), nil)
	if err != nil {
		return err
	}

	taskTemplate3, err := b.TemplateAddWithOutMeta(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates("RCA", "Document the root cause analysis", b.StatusItemOpened.ID, taskEntity), nil)
	if err != nil {
		return err
	}

	cp1 := &base.CoreWorkflow{
		Name:    "Basic Incidents",
		ActorID: incidentEntity.ID,
		Exp:     "",
		Nodes: []*base.CoreNode{
			{
				Name:      "Open",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "open",
				Nodes:     []*base.CoreNode{},
				Exp:       "open",
			},
			{
				Name:      "Triggered",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "triggered",
				Nodes:     []*base.CoreNode{},
				Exp:       "triggered",
			},
			{
				Name:      "Acknowledged",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "acknowledged",
				Nodes:     []*base.CoreNode{},
				Exp:       "acknowledged",
			},
			{
				Name:      "Resolved",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "resolved",
				Nodes: []*base.CoreNode{
					{
						Name:       "Document the incident",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate1.ID,
					},
				},
				Exp: "resolved",
			},
		},
	}

	err = b.AddPipelines(ctx, cp1)
	if err != nil {
		return err
	}

	cp2 := &base.CoreWorkflow{
		Name:    "Critical Incidents",
		ActorID: incidentEntity.ID,
		Exp:     "",
		Nodes: []*base.CoreNode{
			{
				Name:      "Open",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "open",
				Nodes:     []*base.CoreNode{},
				Exp:       "open",
			},
			{
				Name:      "Triggered",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "triggered",
				Nodes: []*base.CoreNode{
					{
						Name:       "Notify managers",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate2.ID,
					},
				},
				Exp: "triggered",
			},
			{
				Name:      "Acknowledged",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "acknowledged",
				Nodes:     []*base.CoreNode{},
				Exp:       "acknowledged",
			},
			{
				Name:      "Resolved",
				ActorID:   "00000000-0000-0000-0000-000000000000",
				ActorName: "resolved",
				Nodes: []*base.CoreNode{
					{
						Name:       "Document RCA",
						ActorID:    taskEntity.ID,
						ActorName:  "Task",
						TemplateID: taskTemplate3.ID,
					},
				},
				Exp: "resolved",
			},
		},
	}

	err = b.AddPipelines(ctx, cp2)
	if err != nil {
		return err
	}

	fmt.Println("\tINCIDENT:AUTOMATION Pipeline And Its Nodes Created")

	return nil
}
