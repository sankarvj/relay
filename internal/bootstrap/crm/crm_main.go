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
	b.LoadFixedEntities(ctx)

	// add entity - deal
	dealEntity, err := b.EntityAdd(ctx, uuid.New().String(), schema.SeedDealsEntityName, "Deals", entity.CategoryData, entity.StateTeamLevel, DealFields(b.ContactEntity.ID, b.CompanyEntity.ID, b.FlowEntity.ID, b.NodeEntity.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Deals Entity Created")

	// add entity - task
	_, err = b.EntityAdd(ctx, uuid.New().String(), schema.SeedTasksEntityName, "Tasks", entity.CategoryTask, entity.StateTeamLevel, TaskFields(b.ContactEntity.ID, b.CompanyEntity.ID, dealEntity.ID, b.StatusEntity.ID, b.NodeEntity.ID, b.StatusItemOpened.ID, b.StatusItemClosed.ID, b.StatusItemOverDue.ID, b.TypeEntity.ID, b.TypeItemEmail.ID, b.TypeItemTodo.ID, b.EmailsEntity.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Tasks Entity Created")

	// add entity - notes
	_, err = b.EntityAdd(ctx, uuid.New().String(), schema.SeedNotesEntityName, "Notes", entity.CategoryNotes, entity.StateTeamLevel, NoteFields(b.ContactEntity.ID, b.CompanyEntity.ID, dealEntity.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Notes Entity Created")

	// add entity - meetings
	_, err = b.EntityAdd(ctx, uuid.New().String(), schema.SeedMeetingsEntityName, "Meetings", entity.CategoryMeeting, entity.StateTeamLevel, MeetingFields(b.ContactEntity.ID, b.CompanyEntity.ID, dealEntity.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Meetings Entity Created")

	// add entity - tickets
	_, err = b.EntityAdd(ctx, uuid.New().String(), schema.SeedTicketsEntityName, "Tickets", entity.CategoryData, entity.StateTeamLevel, base.TicketFields(b.ContactEntity.ID, b.CompanyEntity.ID, b.StatusEntity.ID))
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
	streamEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityStream)
	if err != nil {
		return err
	}

	fmt.Println("\tCRM:SAMPLES All CRM Entities Retrived")

	assID1, assID2, err := b.AddAssociations(ctx, contactEntity, companyEntity, dealEntity, ticketEntity, emailsEntity, streamEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Sample Web Of Associations Created Between All The Above Entities")

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
	contactItem1, err := b.ItemAdd(ctx, contactEntity.ID, uuid.New().String(), b.UserID, forms.ContactVals("Bruce Wayne", "gaajidurden@gmail.com"))
	if err != nil {
		return err
	}
	// add contact item - senthil (straight)
	contactItem2, err := b.ItemAdd(ctx, contactEntity.ID, uuid.New().String(), b.UserID, forms.ContactVals("George Kutty", "vijayasankarmobile@gmail.com"))
	if err != nil {
		return err
	}

	fmt.Println("\tCRM:SAMPLES Contacts Items Created")

	companyItem1, err := b.ItemAddGenie(ctx, companyEntity.ID, uuid.New().String(), b.UserID, base.UUIDHolder, forms.CompanyVals("Zoho", "zoho.com"), map[string]string{contactEntity.ID: contactItem1.ID})
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

	// add delay item
	delayItem, err := b.ItemAdd(ctx, delayEntity.ID, uuid.New().String(), b.UserID, base.DelayVals())
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Delay Item Created")

	pID, _, err := b.AddCRMPipelines(ctx, dealEntity.ID, webhookEntity.ID, delayEntity.ID, delayItem.ID)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Pipeline And Its Nodes Created")

	// _, _, err = b.AddCRMWorkflows1(ctx, contactEntity.ID, taskEntity.ID)
	// if err != nil {
	// 	return err
	// }
	// _, _, err = b.AddCRMWorkflows2(ctx, dealEntity.ID, taskEntity.ID)
	// if err != nil {
	// 	return err
	// }
	fmt.Println("\tCRM:SAMPLES Workflows And Its Nodes Created")

	// add deal item with contacts - vijay & senthil (reverse) & pipeline stage
	dealItem1, err := b.ItemAdd(ctx, dealEntity.ID, uuid.New().String(), b.UserID, DealVals("Big Deal", 1000, contactItem1.ID, contactItem2.ID, pID))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Deal Item Created")

	ticketItem1, err := b.ItemAddGenie(ctx, ticketEntity.ID, uuid.New().String(), b.UserID, base.UUIDHolder, base.TicketVals("My Laptop Is Not Working", statusItems[0].ID), map[string]string{dealEntity.ID: dealItem1.ID})
	if err != nil {
		return err
	}

	fmt.Println("\tCRM:SAMPLES Tickets Items Created")

	// add email-config & email-templates
	err = b.AddEmails(ctx, contactEntity.ID, contactEntity.Key("email"), contactEntity.Key("nps_score"))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES  Email Config Entity And It's Item Created")

	err = b.AddLayouts(ctx, "card", companyEntity.ID)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Sample Layouts Created For All The Above Entities")

	err = b.AddConnections(ctx, assID1, assID2, contactEntity, companyEntity, dealEntity, ticketEntity, contactItem1, companyItem1, dealItem1, ticketItem1)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Sample Web Of Connections Created Between All The Above Items")

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

func AddCompanies(ctx context.Context, b *base.Base) error {

	companyEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, schema.SeedCompaniesEntityName)
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals("Freshworks", "freshworks.com"))
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals("Acme Intl", "acme.com"))
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals("Tesla Inc", "tesla.com"))
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals("Cisco Inc", "cisco.com"))
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals("Salesforce Inc", "salesforce.com"))
	if err != nil {
		return err
	}

	return nil
}
