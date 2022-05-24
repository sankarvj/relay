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

	err = b.AddCSMPipeline(ctx, projectEntity.ID, "Demo Pipeline", "Kick Off", "Solutioning & Traning", "Go-live")
	if err != nil {
		return err
	}
	err = b.AddCSMPipeline(ctx, projectEntity.ID, "Upscale Pipeline", "Meeting", "Product overview", "Traning and Go-live")
	if err != nil {
		return err
	}
	fmt.Println("\tCSM:SAMPLES Pipeline And Its Nodes Created")

	err = b.AddSegments(ctx, projectEntity.ID)
	if err != nil {
		return err
	}
	return nil
}
