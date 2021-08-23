package bootstrap

import (
	"context"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/schema"
)

func BootCRM(db *sqlx.DB, rp *redis.Pool, accountID string) error {
	fmt.Printf("CRM Bootstrap request received for accountID %s\n", accountID)

	ctx := context.Background()
	fmt.Println("\tDB successfully Initialized")

	sampleUserID := UUIDHolder
	// Initialize Team
	teamID := uuid.New().String()
	err := BootstrapTeam(ctx, db, accountID, teamID, "CRM")
	if err != nil {
		return errors.Wrap(err, "account inserted but team bootstrap failed")
	}
	fmt.Println("\tTeam Added")

	// Retrive Owner Entity, Which is created for the account
	ownerEntity, err := entity.RetrieveFixedEntity(ctx, db, accountID, entity.FixedEntityOwner)
	if err != nil {
		return err
	}
	// Retrive Emails Entity, Which is created for the account
	emailsEntity, err := entity.RetrieveFixedEntity(ctx, db, accountID, entity.FixedEntityEmails)
	if err != nil {
		return err
	}
	// Retrive Event Entity, Which is created for the account
	eventEntity, err := entity.RetrieveFixedEntity(ctx, db, accountID, entity.FixedEntityEvent)
	if err != nil {
		return err
	}

	streamEntity, err := entity.RetrieveFixedEntity(ctx, db, accountID, entity.FixedEntityStream)
	if err != nil {
		return err
	}
	fmt.Println("\tOwner, Emails & Events Entities Retrived")

	// Flow wrapper entity added to facilitate other entities(deals) to reference the flows(pipeline) as the reference fields
	flowEntity, err := EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), schema.SeedFlowEntityName, "Flow", entity.CategoryFlow, FlowFields())
	if err != nil {
		return err
	}

	// Node wrapper entity added to facilitate other entities(deals) to reference the stages(pipeline stage) as the reference fields
	nodeEntity, err := EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), schema.SeedNodeEntityName, "Node", entity.CategoryNode, NodeFields())
	if err != nil {
		return err
	}
	fmt.Println("\tFlow & Node Wrapper Entities Created")

	// add status entity
	statusEntity, err := EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), schema.SeedStatusEntityName, "Status", entity.CategoryChildUnit, StatusFields())
	if err != nil {
		return err
	}
	// add status item - open
	statusItemOpen, err := ItemAdd(ctx, db, rp, accountID, statusEntity.ID, uuid.New().String(), sampleUserID, StatusVals(entity.FuExpNone, "Open", "#fb667e"))
	if err != nil {
		return err
	}
	// add status item - closed
	statusItemClosed, err := ItemAdd(ctx, db, rp, accountID, statusEntity.ID, uuid.New().String(), sampleUserID, StatusVals(entity.FuExpDone, "Closed", "#66fb99"))
	if err != nil {
		return err
	}
	// add status item - overdue
	statusItemOverDue, err := ItemAdd(ctx, db, rp, accountID, statusEntity.ID, uuid.New().String(), sampleUserID, StatusVals(entity.FuExpNeg, "OverDue", "#66fb99"))
	if err != nil {
		return err
	}
	fmt.Println("\tStatus Entity With It's Three Statuses Items Created")

	// add type entity
	typeEntity, err := EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), schema.SeedTypeEntityName, "Type", entity.CategoryChildUnit, TypeFields())
	if err != nil {
		return err
	}
	// add type item - email
	typeItemEmail, err := ItemAdd(ctx, db, rp, accountID, typeEntity.ID, uuid.New().String(), sampleUserID, TypeVals(entity.FuExpNone, "Email"))
	if err != nil {
		return err
	}
	// add type item - todo
	typeItemTodo, err := ItemAdd(ctx, db, rp, accountID, typeEntity.ID, uuid.New().String(), sampleUserID, TypeVals(entity.FuExpNone, "Todo"))
	if err != nil {
		return err
	}
	fmt.Println("\tType Entity With It's Three types Items Created")

	// add entity - contacts
	contactEntity, err := EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), schema.SeedContactsEntityName, "Contacts", entity.CategoryData, ContactFields(statusEntity.ID, ownerEntity.ID, ownerEntity.Key("email")))
	if err != nil {
		return err
	}
	// add contact item - vijay (straight)
	contactItem1, err := ItemAdd(ctx, db, rp, accountID, contactEntity.ID, uuid.New().String(), sampleUserID, ContactVals("Bruce Wayne", "gaajidurden@gmail.com", statusItemOpen.ID))
	if err != nil {
		return err
	}
	// add contact item - senthil (straight)
	contactItem2, err := ItemAdd(ctx, db, rp, accountID, contactEntity.ID, uuid.New().String(), sampleUserID, ContactVals("George Kutty", "vijayasankarmobile@gmail.com", statusItemClosed.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tContact Entity With It's Two Contacts Items Created")

	// add entity - companies
	companyEntity, err := EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), schema.SeedCompaniesEntityName, "Companies", entity.CategoryData, CompanyFields(ownerEntity.ID, ownerEntity.Key("email")))
	if err != nil {
		return err
	}
	companyItem1, err := ItemAdd(ctx, db, rp, accountID, companyEntity.ID, uuid.New().String(), sampleUserID, CompanyVals("Zoho", "zoho.com"))
	if err != nil {
		return err
	}
	fmt.Println("\tCompany Entity With It's One Company Item Created")

	// add entity - api-hook
	webhookEntity, err := EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), schema.SeedWebHookEntityName, "WebHook", entity.CategoryAPI, APIFields())
	if err != nil {
		return err
	}
	fmt.Println("\tWebHook Entity Created")

	// add entity - delay
	delayEntity, err := EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), schema.SeedDelayEntityName, "Delay Timer", entity.CategoryDelay, DelayFields())
	if err != nil {
		return err
	}

	// add delay item
	delayItem, err := ItemAdd(ctx, db, rp, accountID, delayEntity.ID, uuid.New().String(), sampleUserID, DelayVals())
	if err != nil {
		return err
	}
	fmt.Println("\tDelay Entity And It's Item Created")

	// add email-config & email-templates
	err = addEmails(ctx, db, accountID, contactEntity.ID, contactEntity.Key("email"), contactEntity.Key("nps_score"))
	if err != nil {
		return err
	}
	fmt.Println("\tEmail Config Entity And It's Item Created")

	// add entity - deal
	dealEntity, err := EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), schema.SeedDealsEntityName, "Deals", entity.CategoryData, DealFields(contactEntity.ID, companyEntity.ID, flowEntity.ID, nodeEntity.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tDeal Entity Created")

	pID, _, err := addPipelines(ctx, db, accountID, dealEntity.ID, webhookEntity.ID, delayEntity.ID, delayItem.ID)
	if err != nil {
		return err
	}
	fmt.Println("\tPipeline And Its Nodes Created")

	// add deal item with contacts - vijay & senthil (reverse) & pipeline stage
	dealItem1, err := ItemAdd(ctx, db, rp, accountID, dealEntity.ID, uuid.New().String(), sampleUserID, DealVals("Big Deal", 1000, contactItem1.ID, contactItem2.ID, pID))
	if err != nil {
		return err
	}
	fmt.Println("\tDeal Item Created")

	// add entity - task
	taskEntity, err := EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), schema.SeedTasksEntityName, "Tasks", entity.CategoryTask, TaskFields(contactEntity.ID, companyEntity.ID, dealEntity.ID, statusEntity.ID, nodeEntity.ID, statusItemOpen.ID, statusItemClosed.ID, statusItemOverDue.ID, typeEntity.ID, typeItemEmail.ID, typeItemTodo.ID, emailsEntity.ID))
	if err != nil {
		return err
	}
	// add task item for contact - vijay (reverse)
	_, err = ItemAdd(ctx, db, rp, accountID, taskEntity.ID, uuid.New().String(), sampleUserID, TaskVals("An Todo Task", contactItem1.ID, typeItemTodo.ID))
	if err != nil {
		return err
	}
	// add task item for contact - vijay (reverse)
	_, err = ItemAdd(ctx, db, rp, accountID, taskEntity.ID, uuid.New().String(), sampleUserID, TaskVals("An Email Task", contactItem1.ID, typeItemEmail.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tTask Entity With It's Two Task Items Created")

	// add entity - notes
	_, err = EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), schema.SeedNotesEntityName, "Notes", entity.CategoryNotes, NoteFields(contactEntity.ID, companyEntity.ID, dealEntity.ID))
	if err != nil {
		return err
	}

	// add entity - meetings
	_, err = EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), schema.SeedMeetingsEntityName, "Meetings", entity.CategoryMeeting, MeetingFields(contactEntity.ID, companyEntity.ID, dealEntity.ID))
	if err != nil {
		return err
	}

	// add entity - tickets
	ticketEntity, err := EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), schema.SeedTicketsEntityName, "Tickets", entity.CategoryData, TicketFields(contactEntity.ID, companyEntity.ID, statusEntity.ID))
	if err != nil {
		return err
	}

	ticketItem1, err := ItemAdd(ctx, db, rp, accountID, ticketEntity.ID, uuid.New().String(), sampleUserID, TicketVals("My Laptop Is Not Working", statusItemOpen.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tTicket Entity And It's Item Created")

	//Event entity add sample item
	_, err = ItemAddGenie(ctx, db, rp, accountID, eventEntity.ID, uuid.New().String(), sampleUserID, dealItem1.ID, EventVals("My first activity recorded", 2))
	if err != nil {
		return err
	}
	fmt.Println("\tEvent Item Created")

	//Stream entity add sample item
	_, err = ItemAddGenie(ctx, db, rp, accountID, streamEntity.ID, uuid.New().String(), sampleUserID, dealItem1.ID, StreamVals("Deal Closed", "Yahooo", ""))
	if err != nil {
		return err
	}
	_, err = ItemAddGenie(ctx, db, rp, accountID, streamEntity.ID, uuid.New().String(), sampleUserID, dealItem1.ID, StreamVals("Deal Weekly Update", "Closing near the deal", ""))
	if err != nil {
		return err
	}
	_, err = ItemAddGenie(ctx, db, rp, accountID, streamEntity.ID, uuid.New().String(), sampleUserID, dealItem1.ID, StreamVals("New task", "This task needs to be added", ""))
	if err != nil {
		return err
	}
	fmt.Println("\tStream Items Created")

	err = addAssociations(ctx, db, accountID, teamID, contactEntity.ID, companyEntity.ID, dealEntity.ID, ticketEntity.ID, contactItem1.ID, companyItem1.ID, dealItem1.ID, ticketItem1.ID, contactEntity.Key("email"), emailsEntity)
	if err != nil {
		return err
	}
	fmt.Println("\tA Web Of Associations Created Between All The Above Entities")

	err = addLayouts(ctx, db, "card", accountID, companyEntity.ID)
	if err != nil {
		return err
	}
	fmt.Println("\tA Layouts Created For All The Above Entities")

	err = addSegments(ctx, db, accountID, contactEntity.ID)
	if err != nil {
		return err
	}
	err = addSegments(ctx, db, accountID, companyEntity.ID)
	if err != nil {
		return err
	}
	err = addSegments(ctx, db, accountID, dealEntity.ID)
	if err != nil {
		return err
	}
	fmt.Println("\tA Segments Created For Contacts/Companies/Deals")

	fmt.Printf("\n\tCRM Bootstrap Successfull!!!! for the account%s\n", accountID)
	return nil
}
