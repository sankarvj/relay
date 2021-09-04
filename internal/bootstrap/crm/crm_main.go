package crm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/base"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/schema"
)

func Boot(ctx context.Context, b *base.Base) error {

	// Retrive Owner Entity, Which is created for the crm
	ownerEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityOwner)
	if err != nil {
		return err
	}

	//Retrive Email Config entity
	emailConfigEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityEmailConfig)
	if err != nil {
		return err
	}

	fmt.Println("\tCRM:BOOT Retrive Owner & EmailConfig Entities Retrived")

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
	fmt.Println("\tCRM:BOOT Flow & Node Wrapper Entities Created")

	// add entity - api-hook
	_, err = b.EntityAdd(ctx, uuid.New().String(), schema.SeedWebHookEntityName, "WebHook", entity.CategoryAPI, entity.StateTeamLevel, APIFields())
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Delay & WebHook Entity Created")

	// add entity - delay
	_, err = b.EntityAdd(ctx, uuid.New().String(), schema.SeedDelayEntityName, "Delay Timer", entity.CategoryDelay, entity.StateTeamLevel, DelayFields())
	if err != nil {
		return err
	}

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

	// add entity - contacts
	contactEntity, err := b.EntityAdd(ctx, uuid.New().String(), schema.SeedContactsEntityName, "Contacts", entity.CategoryData, entity.StateAccountLevel, ContactFields(statusEntity.ID, ownerEntity.ID, ownerEntity.Key("email")))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Contacts Entity Created")

	// add entity - emails
	emailsEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityEmails, "Emails", entity.CategoryEmail, entity.StateTeamLevel, forms.EmailFields(emailConfigEntity.ID, emailConfigEntity.Key("email"), contactEntity.ID, contactEntity.Key("first_name"), contactEntity.Key("email")))
	if err != nil {
		return err
	}
	// add entity - companies
	companyEntity, err := b.EntityAdd(ctx, uuid.New().String(), schema.SeedCompaniesEntityName, "Companies", entity.CategoryData, entity.StateAccountLevel, CompanyFields(ownerEntity.ID, ownerEntity.Key("email")))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Companies Entity Created")

	// add entity - deal
	dealEntity, err := b.EntityAdd(ctx, uuid.New().String(), schema.SeedDealsEntityName, "Deals", entity.CategoryData, entity.StateTeamLevel, DealFields(contactEntity.ID, companyEntity.ID, flowEntity.ID, nodeEntity.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Deals Entity Created")

	// add entity - task
	_, err = b.EntityAdd(ctx, uuid.New().String(), schema.SeedTasksEntityName, "Tasks", entity.CategoryTask, entity.StateTeamLevel, TaskFields(contactEntity.ID, companyEntity.ID, dealEntity.ID, statusEntity.ID, nodeEntity.ID, statusItemOpen.ID, statusItemClosed.ID, statusItemOverDue.ID, typeEntity.ID, typeItemEmail.ID, typeItemTodo.ID, emailsEntity.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Tasks Entity Created")

	// add entity - notes
	_, err = b.EntityAdd(ctx, uuid.New().String(), schema.SeedNotesEntityName, "Notes", entity.CategoryNotes, entity.StateTeamLevel, NoteFields(contactEntity.ID, companyEntity.ID, dealEntity.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Notes Entity Created")

	// add entity - meetings
	_, err = b.EntityAdd(ctx, uuid.New().String(), schema.SeedMeetingsEntityName, "Meetings", entity.CategoryMeeting, entity.StateTeamLevel, MeetingFields(contactEntity.ID, companyEntity.ID, dealEntity.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Meetings Entity Created")

	// add entity - tickets
	_, err = b.EntityAdd(ctx, uuid.New().String(), schema.SeedTicketsEntityName, "Tickets", entity.CategoryData, entity.StateTeamLevel, TicketFields(contactEntity.ID, companyEntity.ID, statusEntity.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Tickets Entity Created")

	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityStream, "Streams", entity.CategoryStream, entity.StateTeamLevel, forms.StreamFields())
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Streams Entity Created")

	return nil
}

func AddSamples(ctx context.Context, b *base.Base) error {

	emailsEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityEmails)
	if err != nil {
		return err
	}
	contactEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, schema.SeedContactsEntityName)
	if err != nil {
		return err
	}
	companyEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, schema.SeedCompaniesEntityName)
	if err != nil {
		return err
	}
	dealEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, schema.SeedDealsEntityName)
	if err != nil {
		return err
	}
	taskEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, schema.SeedTasksEntityName)
	if err != nil {
		return err
	}
	ticketEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, schema.SeedTicketsEntityName)
	if err != nil {
		return err
	}
	statusEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, schema.SeedStatusEntityName)
	if err != nil {
		return err
	}
	typeEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, schema.SeedTypeEntityName)
	if err != nil {
		return err
	}
	delayEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, schema.SeedDelayEntityName)
	if err != nil {
		return err
	}
	webhookEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, schema.SeedWebHookEntityName)
	if err != nil {
		return err
	}

	fmt.Println("\tCRM:SAMPLES All CRM Entities Retrived")

	statusItems, err := item.List(ctx, statusEntity.ID, b.DB)
	if err != nil {
		return err
	}
	typeItems, err := item.List(ctx, typeEntity.ID, b.DB)
	if err != nil {
		return err
	}

	fmt.Println("\tCRM:SAMPLES Needed Items Retrived")

	// add contact item - vijay (straight)
	contactItem1, err := b.ItemAdd(ctx, contactEntity.ID, uuid.New().String(), b.UserID, ContactVals("Bruce Wayne", "gaajidurden@gmail.com", statusItems[1].ID))
	if err != nil {
		return err
	}
	// add contact item - senthil (straight)
	contactItem2, err := b.ItemAdd(ctx, contactEntity.ID, uuid.New().String(), b.UserID, ContactVals("George Kutty", "vijayasankarmobile@gmail.com", statusItems[2].ID))
	if err != nil {
		return err
	}

	fmt.Println("\tCRM:SAMPLES Contacts Items Created")

	companyItem1, err := b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, CompanyVals("Zoho", "zoho.com"))
	if err != nil {
		return err
	}

	fmt.Println("\tCRM:SAMPLES Companies Item Created")

	// add task item for contact - vijay (reverse)
	_, err = b.ItemAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, TaskVals("An Todo Task", contactItem1.ID, typeItems[0].ID))
	if err != nil {
		return err
	}
	// add task item for contact - vijay (reverse)
	_, err = b.ItemAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, TaskVals("An Email Task", contactItem1.ID, typeItems[1].ID))
	if err != nil {
		return err
	}

	fmt.Println("\tCRM:SAMPLES Tasks Items Created")

	ticketItem1, err := b.ItemAdd(ctx, ticketEntity.ID, uuid.New().String(), b.UserID, TicketVals("My Laptop Is Not Working", statusItems[0].ID))
	if err != nil {
		return err
	}

	fmt.Println("\tCRM:SAMPLES Tickets Items Created")

	streamEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityStream)
	if err != nil {
		return err
	}

	// add delay item
	delayItem, err := b.ItemAdd(ctx, delayEntity.ID, uuid.New().String(), b.UserID, DelayVals())
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Delay Item Created")

	pID, _, err := b.AddPipelines(ctx, dealEntity.ID, webhookEntity.ID, delayEntity.ID, delayItem.ID)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Pipeline And Its Nodes Created")

	// add deal item with contacts - vijay & senthil (reverse) & pipeline stage
	dealItem1, err := b.ItemAdd(ctx, dealEntity.ID, uuid.New().String(), b.UserID, DealVals("Big Deal", 1000, contactItem1.ID, contactItem2.ID, pID))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Deal Item Created")

	//Stream entity add sample item
	_, err = b.ItemAddGenie(ctx, streamEntity.ID, uuid.New().String(), b.UserID, dealItem1.ID, forms.StreamVals("Deal Closed", "Yahooo", ""))
	if err != nil {
		return err
	}
	_, err = b.ItemAddGenie(ctx, streamEntity.ID, uuid.New().String(), b.UserID, dealItem1.ID, forms.StreamVals("Deal Weekly Update", "Closing near the deal", ""))
	if err != nil {
		return err
	}
	_, err = b.ItemAddGenie(ctx, streamEntity.ID, uuid.New().String(), b.UserID, dealItem1.ID, forms.StreamVals("New task", "This task needs to be added", ""))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Stream Items Created")

	// add email-config & email-templates
	err = b.AddEmails(ctx, contactEntity.ID, contactEntity.Key("email"), contactEntity.Key("nps_score"))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES  Email Config Entity And It's Item Created")

	err = b.AddAssociations(ctx, contactEntity.ID, companyEntity.ID, dealEntity.ID, ticketEntity.ID, emailsEntity.ID, contactItem1.ID, companyItem1.ID, dealItem1.ID, ticketItem1.ID, contactEntity.Key("email"))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Sample Web Of Associations Created Between All The Above Entities")

	err = b.AddLayouts(ctx, "card", companyEntity.ID)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Sample Layouts Created For All The Above Entities")

	err = b.AddSegments(ctx, contactEntity.ID)
	if err != nil {
		return err
	}
	err = b.AddSegments(ctx, companyEntity.ID)
	if err != nil {
		return err
	}
	err = b.AddSegments(ctx, dealEntity.ID)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Sample Segments Created For Contacts/Companies/Deals")
	return nil
}

func AddProps(ctx context.Context, b *base.Base) error {
	_, err := b.EntityAdd(ctx, uuid.New().String(), schema.SeedPageViewEventEntityName, "Page View", entity.CategoryEvent, entity.StateTeamLevel, pageViewEventEntityFields())
	if err != nil {
		return err
	}

	_, err = b.EntityAdd(ctx, uuid.New().String(), schema.SeedActivityEventEntityName, "Activity View", entity.CategoryEvent, entity.StateTeamLevel, activityEventEntityFields())
	if err != nil {
		return err
	}
	return nil
}
