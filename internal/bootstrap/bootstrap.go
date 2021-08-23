package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/connection"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/layout"
	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

const (
	UUIDHolder = "00000000-0000-0000-0000-000000000000"
)

func BootstrapTeam(ctx context.Context, db *sqlx.DB, accountID, teamID, teamName string) error {
	const q = `INSERT INTO teams
		(team_id, account_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := db.ExecContext(
		ctx, q,
		teamID, accountID, teamName, "", time.Now().UTC(), time.Now().UTC().Unix(),
	)
	return err
}

func BootstrapOwnerEntity(ctx context.Context, db *sqlx.DB, rp *redis.Pool, currentUser *user.User, accountID, teamID string) error {
	fields, itemVals := ownerFields(currentUser.ID, *currentUser.Name, *currentUser.Avatar, currentUser.Email)
	// add entity - owners
	ue, err := EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), entity.FixedEntityOwner, "Owners", entity.CategoryUsers, fields)
	if err != nil {
		return err
	}
	// add owner item
	// pass the currentUserID as the itemID. Is it okay to do like that? seems like a anti pattern.
	_, err = ItemAdd(ctx, db, rp, accountID, ue.ID, currentUser.ID, currentUser.ID, itemVals)
	if err != nil {
		return err
	}
	return nil
}

func BootstrapEmailConfigEntity(ctx context.Context, db *sqlx.DB, accountID, teamID string) error {
	coEntityID, coEmail, err := currentOwner(ctx, db, accountID)
	if err != nil {
		return err
	}
	fields := emailConfigFields(coEntityID, coEmail)
	// add entity - email- configs
	_, err = EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), entity.FixedEntityEmailConfig, "Email Integrations", entity.CategoryEmailConfig, fields)

	return err
}

func BootstrapEmailsEntity(ctx context.Context, db *sqlx.DB, accountID, teamID string) error {
	emailConfigEntity, err := entity.RetrieveFixedEntity(ctx, db, accountID, entity.FixedEntityEmailConfig)
	if err != nil {
		return err
	}

	//TODO uuid-00-email needs to be changed
	fields := emailFields(emailConfigEntity.ID, emailConfigEntity.Key("email"), "", "uuid-00-email")
	// add entity - email
	_, err = EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), entity.FixedEntityEmails, "Emails", entity.CategoryEmail, fields)

	return err
}

func BootstrapCalendarEntity(ctx context.Context, db *sqlx.DB, accountID, teamID string) error {
	coEntityID, coEmail, err := currentOwner(ctx, db, accountID)
	if err != nil {
		return err
	}
	fields := calendarFields(coEntityID, coEmail)
	// add entity - calendar
	_, err = EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), entity.FixedEntityCalendar, "Calendar", entity.CategoryCalendar, fields)

	return err
}

func BootstrapMessageEntity(ctx context.Context, db *sqlx.DB, accountID, teamID string) error {
	fields := EventFields()
	// add entity - event
	_, err := EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), entity.FixedEntityEvent, "Events", entity.CategoryEvent, fields)
	if err != nil {
		return err
	}

	fields = StreamFields()
	// add entity - streams
	_, err = EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), entity.FixedEntityStream, "Streams", entity.CategoryStream, fields)
	return err
}

func BootstrapLayout(ctx context.Context, db *sqlx.DB, name, accountID, entityID string, fields map[string]string) error {
	nl := layout.NewLayout{}
	nl.Name = name
	nl.AccountID = accountID
	nl.EntityID = entityID
	nl.Fields = fields

	_, err := layout.Create(ctx, db, nl, time.Now())
	return err
}

func BootstrapSegments(ctx context.Context, db *sqlx.DB, accountID, entityID, name, exp string) error {
	_, err := FlowAdd(ctx, db, accountID, uuid.New().String(), entityID, name, flow.FlowModeSegment, flow.FlowConditionNil, exp)
	if err != nil {
		return err
	}
	return nil
}

func currentOwner(ctx context.Context, db *sqlx.DB, accountID string) (string, string, error) {
	ownerEntity, err := entity.RetrieveFixedEntity(ctx, db, accountID, entity.FixedEntityOwner)
	if err != nil {
		return "", "", err
	}
	ownerFields, err := ownerEntity.Fields()
	if err != nil {
		return "", "", err
	}
	return ownerEntity.ID, entity.NamedKeysMap(ownerFields)["email"], nil
}

func EntityUpdate(ctx context.Context, db *sqlx.DB, accountID, teamID, entityID string, fields []entity.Field) error {
	input, err := json.Marshal(fields)
	if err != nil {
		return err
	}

	fmt.Printf("\t\tEntity '%s' Updated\n", entityID)
	return entity.Update(ctx, db, accountID, entityID, string(input), time.Now())
}

func EntityAdd(ctx context.Context, db *sqlx.DB, accountID, teamID, entityID, name, displayName string, category int, fields []entity.Field) (entity.Entity, error) {
	ne := entity.NewEntity{
		ID:          entityID,
		AccountID:   accountID,
		TeamID:      teamID,
		Category:    category,
		Name:        name,
		DisplayName: displayName,
		Fields:      fields,
	}

	e, err := entity.Create(ctx, db, ne, time.Now())
	if err != nil {
		return entity.Entity{}, err
	}

	fmt.Printf("\t\tEntity '%s' Bootstraped\n", e.DisplayName)
	return e, nil
}

func ItemAdd(ctx context.Context, db *sqlx.DB, rp *redis.Pool, accountID, entityID, itemID, userID string, fields map[string]interface{}) (item.Item, error) {
	return ItemAddGenie(ctx, db, rp, accountID, entityID, itemID, userID, UUIDHolder, fields)
}

func ItemAddGenie(ctx context.Context, db *sqlx.DB, rp *redis.Pool, accountID, entityID, itemID, userID, genieID string, fields map[string]interface{}) (item.Item, error) {
	ni := item.NewItem{
		ID:        itemID,
		AccountID: accountID,
		EntityID:  entityID,
		UserID:    &userID,
		GenieID:   &genieID,
		Fields:    fields,
	}

	it, err := item.Create(ctx, db, ni, time.Now())
	if err != nil {
		return item.Item{}, err
	}

	j := job.Job{}
	j.EventItemCreated(accountID, entityID, it.ID, ni.Source, db, rp)

	fmt.Printf("\t\t\tItem Added\n")
	return it, nil
}

func FlowAdd(ctx context.Context, db *sqlx.DB, accountID, flowID, entityID string, name string, mode, condition int, exp string) (flow.Flow, error) {
	nf := flow.NewFlow{
		ID:         flowID,
		AccountID:  accountID,
		EntityID:   entityID,
		Mode:       mode,
		Type:       flow.FlowTypeUnknown,
		Condition:  condition,
		Expression: exp,
		Name:       name,
	}

	f, err := flow.Create(ctx, db, nf, time.Now())
	if err != nil {
		return flow.Flow{}, err
	}

	fmt.Printf("\t\tFlow '%s' Bootstraped\n", name)
	return f, nil
}

func NodeAdd(ctx context.Context, db *sqlx.DB, accountID, nodeID, flowID, actorID string, pnodeID string, name string, typ int, exp string, actuals map[string]string, stageID, description string) (node.Node, error) {
	nn := node.NewNode{
		ID:           nodeID,
		AccountID:    accountID,
		FlowID:       flowID,
		ActorID:      actorID,
		ParentNodeID: pnodeID,
		StageID:      stageID,
		Name:         name,
		Description:  description,
		Type:         typ,
		Expression:   exp,
		Actuals:      actuals,
	}

	n, err := node.Create(ctx, db, nn, time.Now())
	if err != nil {
		return node.Node{}, err
	}

	fmt.Printf("\t\t\tNode '%s' Added For Flow %s\n", name, flowID)
	return n, nil
}

func AssociationAdd(ctx context.Context, db *sqlx.DB, accountID, srcEntityID, dstEntityID string) (string, error) {
	relationshipID, err := relationship.Associate(ctx, db, accountID, srcEntityID, dstEntityID)
	if err != nil {
		return "", err
	}

	fmt.Printf("\t\tAssociation added between entities '%s' and '%s'\n", srcEntityID, dstEntityID)
	return relationshipID, nil
}

func ConnectionAdd(ctx context.Context, db *sqlx.DB, accountID, relationshipID, srcItemID, dstItemID string) error {
	err := connection.Associate(ctx, db, accountID, relationshipID, srcItemID, dstItemID)
	if err != nil {
		return err
	}
	fmt.Printf("\t\t\tConnection added between items '%s' and '%s' for the relationship '%s'\n", srcItemID, dstItemID, relationshipID)
	return nil
}
