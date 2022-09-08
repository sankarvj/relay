package crm

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/base"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func CreateContactCompanyTaskEntity(ctx context.Context, b *base.Base) (*entity.Entity, *entity.Entity, *entity.Entity, error) {

	contactEntity, err := entity.RetrieveFixedEntityAccountLevel(ctx, b.DB, b.AccountID, entity.FixedEntityContacts)
	if err == entity.ErrFixedEntityNotFound {
		// add entity - contacts
		conForms := forms.ContactFields(b.OwnerEntity.ID, b.OwnerEntity.Key("email"))
		contactEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityContacts, "Contacts", entity.CategoryData, entity.StateTeamLevel, false, true, true, conForms)
		if err != nil {
			return nil, nil, nil, err
		}
	} else if err == nil {
		// update entity - contacts with crm team-id
		contactEntity.SharedTeamIds = append(contactEntity.SharedTeamIds, b.TeamID)
		err = entity.UpdateSharedTeam(ctx, b.DB, b.AccountID, contactEntity.ID, contactEntity.SharedTeamIds, time.Now())
		if err != nil {
			return nil, nil, nil, err
		}
	}
	companyEntity, err := entity.RetrieveFixedEntityAccountLevel(ctx, b.DB, b.AccountID, entity.FixedEntityCompanies)
	if err == entity.ErrFixedEntityNotFound {
		// add entity - companies
		comForms := forms.CompanyFields(b.OwnerEntity.ID, b.OwnerEntity.Key("email"))
		companyEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityCompanies, "Companies", entity.CategoryData, entity.StateTeamLevel, false, true, true, comForms)
		if err != nil {
			return nil, nil, nil, err
		}
	} else if err == nil {
		// update entity - companies with crm team-id
		companyEntity.SharedTeamIds = append(companyEntity.SharedTeamIds, b.TeamID)
		err = entity.UpdateSharedTeam(ctx, b.DB, b.AccountID, companyEntity.ID, companyEntity.SharedTeamIds, time.Now())
		if err != nil {
			return nil, nil, nil, err
		}
	}

	taskEntity, err := entity.RetrieveFixedEntityAccountLevel(ctx, b.DB, b.AccountID, entity.FixedEntityTask)
	if err == entity.ErrFixedEntityNotFound {
		// add entity - task
		fields := forms.TaskFields(contactEntity.ID, contactEntity.Key("first_name"), companyEntity.ID, companyEntity.Key("name"), b.NodeEntity.ID, b.StatusEntity.ID, b.StatusEntity.Key("name"), b.OwnerEntity.ID, b.OwnerEntity.Key("name"))
		taskEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityTask, "Tasks", entity.CategoryTask, entity.StateTeamLevel, false, false, true, fields)
		if err != nil {
			return nil, nil, nil, err
		}

	} else if err == nil {
		// update entity - tasks with crm team-id
		taskEntity.SharedTeamIds = append(taskEntity.SharedTeamIds, b.TeamID)
		err = entity.UpdateSharedTeam(ctx, b.DB, b.AccountID, taskEntity.ID, taskEntity.SharedTeamIds, time.Now())
		if err != nil {
			return nil, nil, nil, err
		}
	}

	fmt.Println("\tCRM:BOOT Contacts/Companies/Task Entity Created/Update")

	return &contactEntity, &companyEntity, &taskEntity, nil
}

func Boot(ctx context.Context, b *base.Base) error {
	b.LoadFixedEntities(ctx)

	conE, comE, _, err := CreateContactCompanyTaskEntity(ctx, b)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT ConComTask Entity Created")

	// add entity - deal
	dealEntity, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityDeals, "Deals", entity.CategoryData, entity.StateTeamLevel, false, true, false, DealFields(conE.ID, conE.Key("first_name"), comE.ID, comE.Key("name"), b.FlowEntity.ID, b.NodeEntity.ID, b.NodeEntity.Key("node_id")))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Deals Entity Created")

	// add entity - notes
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityNote, "Notes", entity.CategoryNotes, entity.StateTeamLevel, false, false, false, NoteFields(conE.ID, conE.Key("first_name"), comE.ID, comE.Key("name"), dealEntity.ID, dealEntity.Key("deal_name")))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Notes Entity Created")

	// add entity - meetings
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityMeetings, "Meetings", entity.CategoryMeeting, entity.StateTeamLevel, false, false, false, MeetingFields(conE.ID, comE.ID, dealEntity.ID, conE.Key("email"), conE.Key("first_name"), comE.Key("name"), dealEntity.Key("deal_name")))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Meetings Entity Created")

	// add entity - tickets
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityTickets, "Tickets", entity.CategoryData, entity.StateTeamLevel, false, true, false, TicketFields(conE.ID, conE.Key("first_name"), comE.ID, comE.Key("name"), b.StatusEntity.ID, b.StatusEntity.Key("name")))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Tickets Entity Created")

	return nil
}

func AddSamples(ctx context.Context, b *base.Base) error {

	emailsEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityEmails)
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
	taskEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityTask)
	if err != nil {
		return err
	}
	ticketEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityTickets)
	if err != nil {
		return err
	}
	statusEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityStatus)
	if err != nil {
		return err
	}

	streamEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityStream)
	if err != nil {
		return err
	}

	fmt.Println("\tCRM:SAMPLES All CRM Entities Retrived")

	assID1, assID2, err := AddAssociations(ctx, b, contactEntity, companyEntity, dealEntity, ticketEntity, emailsEntity, streamEntity, taskEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Sample Web Of Associations Created Between All The Above Entities")

	statusItems, err := item.List(ctx, b.AccountID, statusEntity.ID, b.DB)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Needed Items Retrived")

	// add contact item
	contactItem1, err := b.ItemAdd(ctx, contactEntity.ID, uuid.New().String(), b.UserID, forms.ContactVals(contactEntity, "Matt Murdock", "matt@starkindst.com"), nil)
	if err != nil {
		return err
	}
	// add contact item
	contactItem2, err := b.ItemAdd(ctx, contactEntity.ID, uuid.New().String(), b.UserID, forms.ContactVals(contactEntity, "Natasha Romanova", "natasha@randcorp.com"), nil)
	if err != nil {
		return err
	}

	contactItem3, err := b.ItemAdd(ctx, contactEntity.ID, uuid.New().String(), b.UserID, forms.ContactVals(contactEntity, "Bruce Banner", "bruce@alumina.com"), nil)
	if err != nil {
		return err
	}

	contactItem4, err := b.ItemAdd(ctx, contactEntity.ID, uuid.New().String(), b.UserID, forms.ContactVals(contactEntity, "Bucky Barnes", "bucky@dailybugle.com"), nil)
	if err != nil {
		return err
	}

	fmt.Println("\tCRM:SAMPLES Contacts Items Created")

	companyItem1, err := b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals(companyEntity, "Stark Industries", "starkindst.com"), map[string][]string{contactEntity.ID: {contactItem1.ID}})
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals(companyEntity, "Rand corporation", "randcorp.com"), map[string][]string{contactEntity.ID: {contactItem2.ID}})
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals(companyEntity, "Alumina", "alumina.com"), map[string][]string{contactEntity.ID: {contactItem3.ID}})
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals(companyEntity, "Daily bugle", "dailybugle.com"), map[string][]string{contactEntity.ID: {contactItem4.ID}})
	if err != nil {
		return err
	}

	fmt.Println("\tCRM:SAMPLES Companies Item Created")

	// add task item for contact - vijay (reverse)
	_, err = b.ItemAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, TaskVals(taskEntity, "An Todo Task", contactItem1.ID), map[string][]string{contactEntity.ID: {contactItem1.ID}})
	if err != nil {
		return err
	}
	// add task item for contact - vijay (reverse)
	_, err = b.ItemAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, TaskVals(taskEntity, "An Email Task", contactItem1.ID), map[string][]string{contactEntity.ID: {contactItem1.ID}})
	if err != nil {
		return err
	}

	fmt.Println("\tCRM:SAMPLES Tasks Items Created")

	err = AddAutomation(ctx, b)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Automations Created")

	// add deal item with contacts - vijay & senthil (reverse) & pipeline stage
	dealItem1, err := b.ItemAdd(ctx, dealEntity.ID, uuid.New().String(), b.UserID, DealVals(dealEntity, "Base Deal", 1000, contactItem1.ID, contactItem2.ID, b.SalesPipelineFlowID), nil)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Deal Item Created")

	ticketItem1, err := b.ItemAdd(ctx, ticketEntity.ID, uuid.New().String(), b.UserID, TicketVals(ticketEntity, "My laptop is not working", statusItems[0].ID), map[string][]string{dealEntity.ID: {dealItem1.ID}})
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Tickets Items Created")

	// add email-config & email-templates
	// err = b.AddEmails(ctx, contactEntity.ID, contactEntity.Key("email"), contactEntity.Key("nps_score"))
	// if err != nil {
	// 	return err
	// }
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

func AddAutomation(ctx context.Context, b *base.Base) error {

	taskEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityTask)
	if err != nil {
		return err
	}

	ticketEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityTickets)
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

	taskTemplate, err := b.TemplateAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, taskTemplates(taskEntity, dealEntity), nil)
	if err != nil {
		return err
	}

	ticketTemplate, err := b.TemplateAdd(ctx, ticketEntity.ID, uuid.New().String(), b.UserID, ticketTemplates(ticketEntity, dealEntity), nil)
	if err != nil {
		return err
	}

	cp := &base.CoreWorkflow{
		Name:    "Sales Pipeline",
		ActorID: dealEntity.ID,
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
					{
						Name:       "Create invoice ticket",
						ActorID:    ticketEntity.ID,
						ActorName:  "Task",
						TemplateID: ticketTemplate.ID,
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

	err = b.AddPipelines(ctx, cp)
	if err != nil {
		return err
	}
	b.SalesPipelineFlowID = cp.FlowID
	fmt.Println("\tCRM:SAMPLES Pipeline And Its Nodes Created")

	// add deal template
	dealTemplate, err := b.TemplateAdd(ctx, dealEntity.ID, uuid.New().String(), b.UserID, dealTemplates(dealEntity, companyEntity, cp.FlowID), nil)
	if err != nil {
		return err
	}

	cf := &base.CoreWorkflow{
		Name:    "When a new company added",
		ActorID: companyEntity.ID,
		Nodes: []*base.CoreNode{
			{
				Name:       "Basic deal",
				ActorID:    dealEntity.ID,
				ActorName:  "Deal",
				TemplateID: dealTemplate.ID,
				Type:       node.Push,
			},
		},
	}

	err = b.AddWorkflows(ctx, cf)
	if err != nil {
		return err
	}

	fmt.Println("\tCRM:SAMPLES Workflows And Its Nodes Created")

	return nil
}

func AddProps(ctx context.Context, b *base.Base) error {
	_, err := b.EntityAdd(ctx, uuid.New().String(), "page_view", "Page View", entity.CategoryEvent, entity.StateTeamLevel, false, false, false, pageViewEventEntityFields())
	if err != nil {
		return err
	}

	_, err = b.EntityAdd(ctx, uuid.New().String(), "activity_view", "Activity View", entity.CategoryEvent, entity.StateTeamLevel, false, false, false, activityEventEntityFields())
	if err != nil {
		return err
	}
	return nil
}

func AddCompanies(ctx context.Context, b *base.Base) error {

	companyEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityCompanies)
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals(companyEntity, "Freshworks", "freshworks.com"), nil)
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals(companyEntity, "Acme Intl", "acme.com"), nil)
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals(companyEntity, "Tesla Inc", "tesla.com"), nil)
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals(companyEntity, "Cisco Inc", "cisco.com"), nil)
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals(companyEntity, "Salesforce Inc", "salesforce.com"), nil)
	if err != nil {
		return err
	}

	return nil
}

func AddAssociations(ctx context.Context, b *base.Base, conEid, comEid, deEid, tickEid, emailEid, streamEID, taskEID entity.Entity) (string, string, error) {
	//contact company association
	assID1, err := b.AssociationAdd(ctx, conEid.ID, comEid.ID)
	if err != nil {
		return "", "", err
	}

	//deal ticket  association
	assID2, err := b.AssociationAdd(ctx, deEid.ID, tickEid.ID)
	if err != nil {
		return "", "", err
	}

	//deal email association
	_, err = b.AssociationAdd(ctx, deEid.ID, emailEid.ID)
	if err != nil {
		return "", "", err
	}

	//deal task association
	_, err = b.AssociationAdd(ctx, deEid.ID, taskEID.ID)
	if err != nil {
		return "", "", err
	}

	//contact email association
	_, err = b.AssociationAdd(ctx, conEid.ID, emailEid.ID)
	if err != nil {
		return "", "", err
	}

	//ticket email association
	_, err = b.AssociationAdd(ctx, tickEid.ID, emailEid.ID)
	if err != nil {
		return "", "", err
	}

	//ASSOCIATE STREAMS
	//contact stream association
	_, err = b.AssociationAdd(ctx, conEid.ID, streamEID.ID)
	if err != nil {
		return "", "", err
	}

	//company stream association
	_, err = b.AssociationAdd(ctx, comEid.ID, streamEID.ID)
	if err != nil {
		return "", "", err
	}

	//deal stream association
	_, err = b.AssociationAdd(ctx, deEid.ID, streamEID.ID)
	if err != nil {
		return "", "", err
	}

	//ticket stream association
	_, err = b.AssociationAdd(ctx, tickEid.ID, streamEID.ID)
	if err != nil {
		return "", "", err
	}

	return assID1, assID2, nil
}
