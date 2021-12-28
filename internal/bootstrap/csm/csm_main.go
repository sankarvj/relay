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

	ownerEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityOwner)
	if err != nil {
		return err
	}

	//Retrive Email Config entity
	emailConfigEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityEmailConfig)
	if err != nil {
		return err
	}
	fmt.Println("\tCSM:BOOT Retrive EmailConfig Entities Retrived")

	contactEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, schema.SeedContactsEntityName)
	if err != nil {
		return err
	}
	companyEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, schema.SeedCompaniesEntityName)
	if err != nil {
		return err
	}
	fmt.Println("\tCSM:BOOT Company & Contacts Entities Retrived")

	// add entity - emails
	emailsEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityEmails, "Emails", entity.CategoryEmail, entity.StateTeamLevel, forms.EmailFields(emailConfigEntity.ID, emailConfigEntity.Key("email"), contactEntity.ID, companyEntity.ID, contactEntity.Key("first_name"), contactEntity.Key("email")))
	if err != nil {
		return err
	}

	// Flow wrapper entity added to facilitate other entities(deals) to reference the flows(pipeline) as the reference fields
	flowEntity, err := b.EntityAdd(ctx, uuid.New().String(), schema.SeedFlowEntityName, "Flow", entity.CategoryFlow, entity.StateTeamLevel, FlowFields())
	if err != nil {
		return err
	}

	// Node wrapper entity added to facilitate other entities(deals) to reference the stages(pipeline stage) as the reference fields
	nodeEntity, err := b.EntityAdd(ctx, uuid.New().String(), schema.SeedNodeEntityName, "Node", entity.CategoryNode, entity.StateTeamLevel, NodeFields())
	if err != nil {
		return err
	}
	fmt.Println("\tCSM:BOOT Flow & Node Wrapper Entities Created")

	// add status entity
	statusEntity, err := b.EntityAdd(ctx, uuid.New().String(), schema.SeedStatusEntityName, "Status", entity.CategoryChildUnit, entity.StateTeamLevel, StatusFields())
	if err != nil {
		return err
	}
	// add status item - open
	statusItemOpen, err := b.ItemAdd(ctx, statusEntity.ID, uuid.New().String(), b.UserID, StatusVals(entity.FuExpNone, "Open", "#fb667e"))
	if err != nil {
		return err
	}
	// add status item - closed
	statusItemClosed, err := b.ItemAdd(ctx, statusEntity.ID, uuid.New().String(), b.UserID, StatusVals(entity.FuExpDone, "Closed", "#66fb99"))
	if err != nil {
		return err
	}
	// add status item - overdue
	statusItemOverDue, err := b.ItemAdd(ctx, statusEntity.ID, uuid.New().String(), b.UserID, StatusVals(entity.FuExpNeg, "OverDue", "#66fb99"))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Status Entity With It's Three Statuses Items Created")

	// add type entity
	typeEntity, err := b.EntityAdd(ctx, uuid.New().String(), schema.SeedTypeEntityName, "Type", entity.CategoryChildUnit, entity.StateTeamLevel, TypeFields())
	if err != nil {
		return err
	}
	// add type item - email
	typeItemEmail, err := b.ItemAdd(ctx, typeEntity.ID, uuid.New().String(), b.UserID, TypeVals(entity.FuExpNone, "Email"))
	if err != nil {
		return err
	}
	// add type item - todo
	typeItemTodo, err := b.ItemAdd(ctx, typeEntity.ID, uuid.New().String(), b.UserID, TypeVals(entity.FuExpNone, "Todo"))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Type Entity With It's Three types Items Created")

	// add entity - project
	projectEntity, err := b.EntityAdd(ctx, uuid.New().String(), schema.SeedProjectsEntityName, "Projects", entity.CategoryData, entity.StateTeamLevel, ProjectFields(statusEntity.ID, ownerEntity.ID, ownerEntity.Key("email"), contactEntity.ID, companyEntity.ID, flowEntity.ID, nodeEntity.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tCSM:BOOT Deals Entity Created")

	// add entity - task
	_, err = b.EntityAdd(ctx, uuid.New().String(), schema.SeedTasksEntityName, "Tasks", entity.CategoryTask, entity.StateTeamLevel, TaskFields(contactEntity.ID, companyEntity.ID, projectEntity.ID, statusEntity.ID, nodeEntity.ID, statusItemOpen.ID, statusItemClosed.ID, statusItemOverDue.ID, typeEntity.ID, typeItemEmail.ID, typeItemTodo.ID, emailsEntity.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tCSM:BOOT Tasks Entity Created")

	// add entity - meetings
	_, err = b.EntityAdd(ctx, uuid.New().String(), schema.SeedMeetingsEntityName, "Meetings", entity.CategoryMeeting, entity.StateTeamLevel, MeetingFields(contactEntity.ID, companyEntity.ID, projectEntity.ID))
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
