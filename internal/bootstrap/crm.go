package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/schema"
)

func BootCRM(cfg database.Config, secDB database.SecConfig, accountID string) error {
	fmt.Printf("CRM Bootstrap request received for accountID %s\n", accountID)

	// Initialize DB
	db, err := database.Open(cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	// Initialize Redis DB
	redisPool := &redis.Pool{
		MaxIdle:     50,
		MaxActive:   50,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", secDB.Host)
			if err != nil {
				return nil, err
			}
			return c, err
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	defer redisPool.Close()
	ctx := context.Background()
	fmt.Println("\tDB successfully Initialized")

	// Initialize Team
	teamID := uuid.New().String()
	err = BootstrapTeam(ctx, db, accountID, teamID, "CRM")
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
	fmt.Println("\tOwner & Emails Entities Retrived")

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
	statusItemOpen, err := ItemAdd(ctx, db, redisPool, accountID, statusEntity.ID, uuid.New().String(), StatusVals(entity.FuExpNone, "Open", "#fb667e"))
	if err != nil {
		return err
	}
	// add status item - closed
	statusItemClosed, err := ItemAdd(ctx, db, redisPool, accountID, statusEntity.ID, uuid.New().String(), StatusVals(entity.FuExpDone, "Closed", "#66fb99"))
	if err != nil {
		return err
	}
	// add status item - overdue
	statusItemOverDue, err := ItemAdd(ctx, db, redisPool, accountID, statusEntity.ID, uuid.New().String(), StatusVals(entity.FuExpNeg, "OverDue", "#66fb99"))
	if err != nil {
		return err
	}
	fmt.Println("\tStatus Entity With It's Three Statuses Items Created")

	// add entity - contacts
	contactEntity, err := EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), schema.SeedContactsEntityName, "Contacts", entity.CategoryData, ContactFields(statusEntity.ID, ownerEntity.ID, ownerEntity.Key("email")))
	if err != nil {
		return err
	}
	// add contact item - vijay (straight)
	contactItem1, err := ItemAdd(ctx, db, redisPool, accountID, contactEntity.ID, uuid.New().String(), ContactVals("Bruce Wayne", "gaajidurden@gmail.com", statusItemOpen.ID))
	if err != nil {
		return err
	}
	// add contact item - senthil (straight)
	contactItem2, err := ItemAdd(ctx, db, redisPool, accountID, contactEntity.ID, uuid.New().String(), ContactVals("George Kutty", "vijayasankarmobile@gmail.com", statusItemClosed.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tContact Entity With It's Two Contacts Items Created")

	// add entity - companies
	companyEntity, err := EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), schema.SeedCompaniesEntityName, "Companies", entity.CategoryData, CompanyFields(ownerEntity.ID, ownerEntity.Key("email")))
	if err != nil {
		return err
	}
	companyItem1, err := ItemAdd(ctx, db, redisPool, accountID, companyEntity.ID, uuid.New().String(), CompanyVals("Zoho", "zoho.com"))
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
	delayItem, err := ItemAdd(ctx, db, redisPool, accountID, delayEntity.ID, uuid.New().String(), DelayVals())
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
	dealItem1, err := ItemAdd(ctx, db, redisPool, accountID, dealEntity.ID, uuid.New().String(), DealVals("Big Deal", 1000, contactItem1.ID, contactItem2.ID, pID))
	if err != nil {
		return err
	}
	fmt.Println("\tDeal Item Created")

	// add entity - task
	taskEntity, err := EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), schema.SeedTasksEntityName, "Tasks", entity.CategoryTask, TaskFields(contactEntity.ID, companyEntity.ID, dealEntity.ID, statusEntity.ID, statusItemOpen.ID, statusItemClosed.ID, statusItemOverDue.ID))
	if err != nil {
		return err
	}
	// add task item for contact - vijay (reverse)
	_, err = ItemAdd(ctx, db, redisPool, accountID, taskEntity.ID, uuid.New().String(), TaskVals("make cake", contactItem1.ID))
	if err != nil {
		return err
	}
	// add task item for contact - vijay (reverse)
	_, err = ItemAdd(ctx, db, redisPool, accountID, taskEntity.ID, uuid.New().String(), TaskVals("make call", contactItem1.ID))
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

	ticketItem1, err := ItemAdd(ctx, db, redisPool, accountID, ticketEntity.ID, uuid.New().String(), TicketVals("My Laptop Is Not Working", statusItemOpen.ID))
	if err != nil {
		return err
	}
	fmt.Println("\tTicket Entity And It's Item Created")

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

	fmt.Printf("\n\tCRM Bootstrap Successfull!!!! for the account%s\n", accountID)
	return nil
}
