package crm

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/base"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func CreateContactCompanyTaskEntity(ctx context.Context, b *base.Base) error {
	var err error
	leadStatusEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityLeadStatus)
	if err != nil {
		return err
	}

	companyEntity, err := entity.RetrieveFixedEntityAccountLevel(ctx, b.DB, b.AccountID, entity.FixedEntityCompanies)
	if err == entity.ErrFixedEntityNotFound {
		// add entity - companies
		comForms := forms.CompanyFields(b.OwnerEntity.ID, b.OwnerEntity.Key("email"))
		companyEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityCompanies, "Companies", entity.CategoryData, entity.StateTeamLevel, false, true, true, comForms)
		if err != nil {
			return err
		}
	} else if err == nil {
		// update entity - companies with crm team-id
		companyEntity.SharedTeamIds = append(companyEntity.SharedTeamIds, b.TeamID)
		err = entity.UpdateSharedTeam(ctx, b.DB, b.AccountID, companyEntity.ID, companyEntity.SharedTeamIds, time.Now())
		if err != nil {
			return err
		}
	}

	contactEntity, err := entity.RetrieveFixedEntityAccountLevel(ctx, b.DB, b.AccountID, entity.FixedEntityContacts)
	if err == entity.ErrFixedEntityNotFound {
		// add entity - contacts
		conForms := forms.ContactFields(b.OwnerEntity.ID, b.OwnerEntity.Key("name"), companyEntity.ID, companyEntity.Key("name"), leadStatusEntity.ID, leadStatusEntity.Key("name"))
		contactEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityContacts, "Contacts", entity.CategoryData, entity.StateTeamLevel, false, true, true, conForms)
		if err != nil {
			return err
		}
	} else if err == nil {
		// update entity - contacts with crm team-id
		contactEntity.SharedTeamIds = append(contactEntity.SharedTeamIds, b.TeamID)
		err = entity.UpdateSharedTeam(ctx, b.DB, b.AccountID, contactEntity.ID, contactEntity.SharedTeamIds, time.Now())
		if err != nil {
			return err
		}
	}

	taskEntity, err := entity.RetrieveFixedEntityAccountLevel(ctx, b.DB, b.AccountID, entity.FixedEntityTask)
	if err == entity.ErrFixedEntityNotFound {
		// add entity - task
		fields := forms.TaskFields(contactEntity.ID, contactEntity.Key("first_name"), companyEntity.ID, companyEntity.Key("name"), b.NodeEntity.ID, b.StatusEntity.ID, b.StatusEntity.Key("name"), b.OwnerEntity.ID, b.OwnerEntity.Key("name"))
		taskEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityTask, "Tasks", entity.CategoryTask, entity.StateTeamLevel, false, false, true, fields)
		if err != nil {
			return err
		}

	} else if err == nil {
		// update entity - tasks with crm team-id
		taskEntity.SharedTeamIds = append(taskEntity.SharedTeamIds, b.TeamID)
		err = entity.UpdateSharedTeam(ctx, b.DB, b.AccountID, taskEntity.ID, taskEntity.SharedTeamIds, time.Now())
		if err != nil {
			return err
		}
	}
	b.ContactEntity = contactEntity
	b.CompanyEntity = companyEntity
	b.TaskEntity = taskEntity

	fmt.Println("\tCRM:BOOT Contacts/Companies/Task Entity Created/Update")

	return nil
}

func Boot(ctx context.Context, b *base.Base) error {
	var err error
	b.LoadFixedEntities(ctx)

	// add entity - lead status
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityLeadStatus, "Lead Status", entity.CategoryChildUnit, entity.StateAccountLevel, false, false, true, forms.LeadStatusFields())
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Lead Status Entity Created")

	err = CreateContactCompanyTaskEntity(ctx, b)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT ConComTask Entity Created")

	// add entity - deal
	b.DealEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityDeals, "Deals", entity.CategoryData, entity.StateTeamLevel, false, true, false, DealFields(b.ContactEntity.ID, b.ContactEntity.Key("first_name"), b.CompanyEntity.ID, b.CompanyEntity.Key("name"), b.FlowEntity.ID, b.NodeEntity.ID, b.NodeEntity.Key("node_id")))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Deals Entity Created")

	// add entity - notes
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityNote, "Notes", entity.CategoryNotes, entity.StateTeamLevel, false, false, false, NoteFields(b.ContactEntity.ID, b.ContactEntity.Key("first_name"), b.CompanyEntity.ID, b.CompanyEntity.Key("name"), b.DealEntity.ID, b.DealEntity.Key("deal_name")))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Notes Entity Created")

	// add entity - meetings
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityMeetings, "Meetings", entity.CategoryMeeting, entity.StateTeamLevel, false, false, false, MeetingFields(b.ContactEntity.ID, b.CompanyEntity.ID, b.DealEntity.ID, b.ContactEntity.Key("email"), b.ContactEntity.Key("first_name"), b.CompanyEntity.Key("name"), b.DealEntity.Key("deal_name")))
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Meetings Entity Created")

	// add entity - tickets
	// _, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityTickets, "Tickets", entity.CategoryData, entity.StateTeamLevel, false, true, false, TicketFields(conE.ID, conE.Key("first_name"), comE.ID, comE.Key("name"), b.StatusEntity.ID, b.StatusEntity.Key("name")))
	// if err != nil {
	// 	return err
	// }
	// fmt.Println("\tCRM:BOOT Tickets Entity Created")

	return nil
}

func AddWorkflows(ctx context.Context, b *base.Base) error {
	var err error
	err = addAutomations(ctx, b)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Automations Created")

	err = b.AddLayouts(ctx, "card", b.CompanyEntity.ID)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Sample Layouts Created For All The Above Entities")

	err = b.AddSegments(ctx, b.ContactEntity.ID)
	if err != nil {
		return err
	}
	err = b.AddSegments(ctx, b.CompanyEntity.ID)
	if err != nil {
		return err
	}
	err = b.AddSegments(ctx, b.DealEntity.ID)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Sample Segments Created For Contacts/Companies/Deals")
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
	leadStatusEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityLeadStatus)
	if err != nil {
		return err
	}
	streamEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityStream)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES All CRM Entities Retrived")

	err = addAssociations(ctx, b, contactEntity, companyEntity, dealEntity, emailsEntity, streamEntity, taskEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Sample Web Of Associations Created Between All The Above Entities")

	err = addLeadStatus(ctx, leadStatusEntity, b)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Lead Status Items Created")

	err = addContacts(ctx, b, contactEntity, taskEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Contacts Items Created")

	err = addCompanies(ctx, b, companyEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Companies Item Created")

	err = addDeals(ctx, b, dealEntity, contactEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Deal Item Created")

	return nil
}

func addAssociations(ctx context.Context, b *base.Base, conEid, comEid, deEid, emailEid, streamEID, taskEID entity.Entity) error {
	//contact company association
	_, err := b.AssociationAdd(ctx, conEid.ID, comEid.ID)
	if err != nil {
		return err
	}

	//deal email association
	_, err = b.AssociationAdd(ctx, deEid.ID, emailEid.ID)
	if err != nil {
		return err
	}

	//deal task association
	_, err = b.AssociationAdd(ctx, deEid.ID, taskEID.ID)
	if err != nil {
		return err
	}

	//contact email association
	_, err = b.AssociationAdd(ctx, conEid.ID, emailEid.ID)
	if err != nil {
		return err
	}

	//ASSOCIATE STREAMS
	//contact stream association
	_, err = b.AssociationAdd(ctx, conEid.ID, streamEID.ID)
	if err != nil {
		return err
	}

	//company stream association
	_, err = b.AssociationAdd(ctx, comEid.ID, streamEID.ID)
	if err != nil {
		return err
	}

	//deal stream association
	_, err = b.AssociationAdd(ctx, deEid.ID, streamEID.ID)
	if err != nil {
		return err
	}

	//ticket stream association
	// _, err = b.AssociationAdd(ctx, tickEid.ID, streamEID.ID)
	// if err != nil {
	// 	return "", "", err
	// }
	//ticket email association
	// _, err = b.AssociationAdd(ctx, tickEid.ID, emailEid.ID)
	// if err != nil {
	// 	return "", "", err
	// }
	//deal ticket  association
	// assID2, err := b.AssociationAdd(ctx, deEid.ID, tickEid.ID)
	// if err != nil {
	// 	return "", "", err
	// }

	return nil
}

func addProps(ctx context.Context, b *base.Base) error {
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

func addContacts(ctx context.Context, b *base.Base, contactEntity, taskEntity entity.Entity) error {
	var err error
	// add contact item
	b.ContactItemMatt, err = b.ItemAdd(ctx, contactEntity.ID, uuid.New().String(), b.UserID, forms.ContactVals(contactEntity, "Matt", "Murdock", "matt@starkindst.com", b.LeadStatusItemNew.ID), nil)
	if err != nil {
		return err
	}
	// add contact item
	b.ContactItemNatasha, err = b.ItemAdd(ctx, contactEntity.ID, uuid.New().String(), b.UserID, forms.ContactVals(contactEntity, "Natasha", "Romanova", "natasha@randcorp.com", b.LeadStatusItemConnected.ID), nil)
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, contactEntity.ID, uuid.New().String(), b.UserID, forms.ContactVals(contactEntity, "Bruce", "Banner", "bruce@alumina.com", b.LeadStatusItemAttempted.ID), nil)
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, contactEntity.ID, uuid.New().String(), b.UserID, forms.ContactVals(contactEntity, "Bucky", "Barnes", "bucky@dailybugle.com", b.LeadStatusItemBadTiming.ID), nil)
	if err != nil {
		return err
	}

	// add task item for contact - matt (reverse)
	_, err = b.ItemAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, TaskVals(taskEntity, "Send demo link to the customer", b.ContactItemMatt.ID), map[string][]string{contactEntity.ID: {b.ContactItemMatt.ID}})
	if err != nil {
		return err
	}
	// add task item for contact - natasha (reverse)
	_, err = b.ItemAdd(ctx, taskEntity.ID, uuid.New().String(), b.UserID, TaskVals(taskEntity, "Schedule an on-site meeting with customer", b.ContactItemNatasha.ID), map[string][]string{contactEntity.ID: {b.ContactItemNatasha.ID}})
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:SAMPLES Tasks Items Created For Matt & Natasha")

	return nil
}

func addCompanies(ctx context.Context, b *base.Base, companyEntity entity.Entity) error {
	var err error
	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals(companyEntity, "Stark Industries", "starkindst.com"), nil)
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals(companyEntity, "Rand corporation", "randcorp.com"), nil)
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals(companyEntity, "Alumina", "alumina.com"), nil)
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals(companyEntity, "Daily bugle", "dailybugle.com"), nil)
	if err != nil {
		return err
	}

	_, err = b.ItemAdd(ctx, companyEntity.ID, uuid.New().String(), b.UserID, forms.CompanyVals(companyEntity, "Salesforce Inc", "salesforce.com"), nil)
	if err != nil {
		return err
	}

	return nil
}

func addLeadStatus(ctx context.Context, leadStatusEntity entity.Entity, b *base.Base) error {
	var err error

	// add status item - new
	b.LeadStatusItemNew, err = b.ItemAdd(ctx, leadStatusEntity.ID, uuid.New().String(), b.UserID, forms.LeadStatusVals(leadStatusEntity, entity.FuExpNone, "New", "#31E1F7"), nil)
	if err != nil {
		return err
	}
	// add status item - open
	_, err = b.ItemAdd(ctx, leadStatusEntity.ID, uuid.New().String(), b.UserID, forms.LeadStatusVals(leadStatusEntity, entity.FuExpNone, "Open", "#7FB77E"), nil)
	if err != nil {
		return err
	}
	// add status item - in-progress
	_, err = b.ItemAdd(ctx, leadStatusEntity.ID, uuid.New().String(), b.UserID, forms.LeadStatusVals(leadStatusEntity, entity.FuExpNone, "In progress", "#FBDF07"), nil)
	if err != nil {
		return err
	}

	// add status item - unqualified
	_, err = b.ItemAdd(ctx, leadStatusEntity.ID, uuid.New().String(), b.UserID, forms.LeadStatusVals(leadStatusEntity, entity.FuExpNeg, "Unqualified", "#FF4A4A"), nil)
	if err != nil {
		return err
	}

	// add status item - attempted to contact
	b.LeadStatusItemAttempted, err = b.ItemAdd(ctx, leadStatusEntity.ID, uuid.New().String(), b.UserID, forms.LeadStatusVals(leadStatusEntity, entity.FuExpNeg, "Attempted to contact", "#781C68"), nil)
	if err != nil {
		return err
	}

	// add status item - bad timing
	b.LeadStatusItemBadTiming, err = b.ItemAdd(ctx, leadStatusEntity.ID, uuid.New().String(), b.UserID, forms.LeadStatusVals(leadStatusEntity, entity.FuExpNeg, "Bad Timing", "#2A0944"), nil)
	if err != nil {
		return err
	}

	// add status item - churned
	_, err = b.ItemAdd(ctx, leadStatusEntity.ID, uuid.New().String(), b.UserID, forms.LeadStatusVals(leadStatusEntity, entity.FuExpNeg, "Chruned", "#2C3333"), nil)
	if err != nil {
		return err
	}

	// add status item - coneected
	b.LeadStatusItemConnected, err = b.ItemAdd(ctx, leadStatusEntity.ID, uuid.New().String(), b.UserID, forms.LeadStatusVals(leadStatusEntity, entity.FuExpPos, "Connected", "#377D71"), nil)
	if err != nil {
		return err
	}
	fmt.Println("\tBOOT:CRM Lead Status Items Created")

	return nil
}

func addDeals(ctx context.Context, b *base.Base, dealEntity, contactEntity entity.Entity) error {
	_, err := b.ItemAdd(ctx, dealEntity.ID, uuid.New().String(), b.UserID, DealVals(dealEntity, "Base Deal", 1000, b.ContactItemMatt.ID, b.ContactItemNatasha.ID, b.SalesPipelineFlowID), nil)
	if err != nil {
		return err
	}
	return nil
}

// func addTickets(ctx context.Context, b *base.Base, dealEntity entity.Entity, dealItem1 item.Item) error {
// 	ticketEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityTickets)
// 	if err != nil {
// 		return err
// 	}
// 	statusEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityStatus)
// 	if err != nil {
// 		return err
// 	}
// 	statusItems, err := item.List(ctx, b.AccountID, statusEntity.ID, b.DB)
// 	if err != nil {
// 		return err
// 	}
// 	_, err = b.ItemAdd(ctx, ticketEntity.ID, uuid.New().String(), b.UserID, TicketVals(ticketEntity, "My laptop is not working", statusItems[0].ID), map[string][]string{dealEntity.ID: {dealItem1.ID}})
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
