package ctm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/base"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/crm"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func Boot(ctx context.Context, b *base.Base) error {
	var err error
	b.LoadFixedEntities(ctx)

	// add entity - ticket status
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityTicketStatus, "Ticket Status", entity.CategoryChildUnit, entity.StateAccountLevel, false, false, true, forms.LeadStatusFields())
	if err != nil {
		return err
	}
	fmt.Println("\tSUPPORT:BOOT Ticket status Status Entity Created")

	err = crm.CreateContactCompanyTaskEntity(ctx, b)
	if err != nil {
		return err
	}
	fmt.Println("\tSUPPORT:BOOT ConComTask Entity Created")

	// add entity - tickets
	b.TicketEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityTickets, "Tickets", entity.CategoryData, entity.StateTeamLevel, false, true, false, forms.TicketFields(b.ContactEntity.ID, b.ContactEntity.Key("first_name"), b.CompanyEntity.ID, b.CompanyEntity.Key("name"), b.StatusEntity.ID, b.StatusEntity.Key("name")))
	if err != nil {
		return err
	}
	fmt.Println("\tSUPPORT:BOOT Tickets Entity Created")

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
	ticketEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityTickets)
	if err != nil {
		return err
	}
	ticketStatusEntity, err := entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityTicketStatus)
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
	fmt.Println("\tSUPPORT:SAMPLES All Support Entities Retrived")

	err = addAssociations(ctx, b, contactEntity, companyEntity, ticketEntity, streamEntity, taskEntity, emailsEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tSUPPORT:SAMPLES Sample Web Of Associations Created Between All The Above Entities")

	err = addTicketStatus(ctx, ticketStatusEntity, b)
	if err != nil {
		return err
	}
	fmt.Println("\tSUPPORT:SAMPLES Ticket Status Items Created")

	err = addTickets(ctx, b, ticketEntity, b.ContactItemMatt.ID, b.TicketStatusItemNew.ID)
	if err != nil {
		return err
	}
	fmt.Println("\tSUPPORT:SAMPLES Ticket Item Created")

	return nil
}

func addAssociations(ctx context.Context, b *base.Base, conEid, comEid, ticEid, streamEID, taskEID, emailEID entity.Entity) error {
	//contact company association
	_, err := b.AssociationAdd(ctx, conEid.ID, comEid.ID)
	if err != nil {
		fmt.Println("ignoring error here. because the contraint might fail if it is alreay added in CRP")
		//return err
	}

	//ticket task association
	_, err = b.AssociationAdd(ctx, ticEid.ID, taskEID.ID)
	if err != nil {
		return err
	}

	//ticket stream association
	_, err = b.AssociationAdd(ctx, ticEid.ID, streamEID.ID)
	if err != nil {
		return err
	}

	//ticket email association
	_, err = b.AssociationAdd(ctx, ticEid.ID, emailEID.ID)
	if err != nil {
		return err
	}

	return nil
}

func addTicketStatus(ctx context.Context, ticketStatusEntity entity.Entity, b *base.Base) error {
	var err error

	// add status item - new
	b.TicketStatusItemNew, err = b.ItemAdd(ctx, ticketStatusEntity.ID, uuid.New().String(), b.UserID, forms.LeadStatusVals(ticketStatusEntity, entity.FuExpNone, "New", "#31E1F7"), nil)
	if err != nil {
		return err
	}
	// add status item - open
	_, err = b.ItemAdd(ctx, ticketStatusEntity.ID, uuid.New().String(), b.UserID, forms.LeadStatusVals(ticketStatusEntity, entity.FuExpNone, "Open", "#7FB77E"), nil)
	if err != nil {
		return err
	}
	// add status item - in-progress
	_, err = b.ItemAdd(ctx, ticketStatusEntity.ID, uuid.New().String(), b.UserID, forms.LeadStatusVals(ticketStatusEntity, entity.FuExpNone, "In progress", "#FBDF07"), nil)
	if err != nil {
		return err
	}

	// add status item - blocked
	_, err = b.ItemAdd(ctx, ticketStatusEntity.ID, uuid.New().String(), b.UserID, forms.LeadStatusVals(ticketStatusEntity, entity.FuExpNeg, "Blocked", "#2C3333"), nil)
	if err != nil {
		return err
	}

	// add status item - closed
	_, err = b.ItemAdd(ctx, ticketStatusEntity.ID, uuid.New().String(), b.UserID, forms.LeadStatusVals(ticketStatusEntity, entity.FuExpPos, "Closed", "#377D71"), nil)
	if err != nil {
		return err
	}
	fmt.Println("\tBOOT:SUPPORT Ticket Status Items Created")

	return nil
}

func addTickets(ctx context.Context, b *base.Base, ticketEntity entity.Entity, contactID, statusNewID string) error {
	_, err := b.ItemAdd(ctx, ticketEntity.ID, uuid.New().String(), b.UserID, TicketVals(ticketEntity, "App crashed when loading home page", contactID, statusNewID), nil)
	if err != nil {
		return err
	}
	return nil
}
