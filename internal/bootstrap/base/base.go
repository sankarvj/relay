package base

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/connection"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/layout"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/schema"
)

const (
	UUIDHolder = "00000000-0000-0000-0000-000000000000"
)

type Base struct {
	AccountID       string
	TeamID          string
	UserID          string
	DB              *sqlx.DB
	RP              *redis.Pool
	FirebaseSDKPath string
	CoreEntity
	CoreItem
	CoreAutomation
}

// These entites must be created or loaded before adding a new product
type CoreEntity struct {
	ContactEntity     entity.Entity
	CompanyEntity     entity.Entity
	OwnerEntity       entity.Entity
	EmailsEntity      entity.Entity
	EmailConfigEntity entity.Entity
	FlowEntity        entity.Entity
	NodeEntity        entity.Entity
	StatusEntity      entity.Entity
	TypeEntity        entity.Entity
}

type CoreItem struct {
	StatusItemOpened  item.Item
	StatusItemClosed  item.Item
	StatusItemOverDue item.Item
	TypeItemEmail     item.Item
	TypeItemTodo      item.Item
}

type CoreAutomation struct {
	SalesPipelineFlowID string
}

type CoreWorkflow struct {
	FlowID  string
	ActorID string
	Name    string
	Exp     string
	Nodes   []*CoreNode
}

type CoreNode struct {
	NodeID     string
	ActorID    string
	ActorName  string
	TemplateID string
	Name       string
	Exp        string
	Nodes      []*CoreNode // nodes inside stages
}

func NewBase(accountID, teamID, userID string, db *sqlx.DB, rp *redis.Pool, firebaseSDKPath string) *Base {
	return &Base{
		AccountID:       accountID,
		TeamID:          teamID,
		UserID:          userID,
		DB:              db,
		RP:              rp,
		FirebaseSDKPath: firebaseSDKPath,
	}
}

func (b *Base) LoadFixedEntities(ctx context.Context) error {
	var err error
	//Retrive Owner Entity
	b.OwnerEntity, err = entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityOwner)
	if err != nil {
		return err
	}

	// Retrive Contact Entity
	b.ContactEntity, err = entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityContacts)
	if err != nil {
		return err
	}

	// Retrive Company Entity
	b.CompanyEntity, err = entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityCompanies)
	if err != nil {
		return err
	}

	//Retrive Email Config entity
	b.EmailConfigEntity, err = entity.RetrieveFixedEntity(ctx, b.DB, b.AccountID, b.TeamID, entity.FixedEntityEmailConfig)
	if err != nil {
		return err
	}

	fmt.Println("\tCRM:BOOT Retrived Owner,Contact,Company & EmailConfig")

	// add entity - emails
	b.EmailsEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityEmails, "Emails", entity.CategoryEmail, entity.StateTeamLevel, forms.EmailFields(b.EmailConfigEntity.ID, b.EmailConfigEntity.Key("email"), b.ContactEntity.ID, b.CompanyEntity.ID, b.ContactEntity.Key("first_name"), b.ContactEntity.Key("email")))
	if err != nil {
		return err
	}

	// Flow wrapper entity added to facilitate other entities(deals) to reference the flows(pipeline) as the reference fields
	b.FlowEntity, err = b.EntityAdd(ctx, uuid.New().String(), schema.SeedFlowEntityName, "Flow", entity.CategoryFlow, entity.StateTeamLevel, FlowFields())
	if err != nil {
		return err
	}

	// Node wrapper entity added to facilitate other entities(deals) to reference the stages(pipeline stage) as the reference fields
	b.NodeEntity, err = b.EntityAdd(ctx, uuid.New().String(), schema.SeedNodeEntityName, "Node", entity.CategoryNode, entity.StateTeamLevel, NodeFields())
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
	b.StatusEntity, err = b.EntityAdd(ctx, uuid.New().String(), schema.SeedStatusEntityName, "Status", entity.CategoryChildUnit, entity.StateTeamLevel, StatusFields())
	if err != nil {
		return err
	}
	// add status item - open
	b.StatusItemOpened, err = b.ItemAdd(ctx, b.StatusEntity.ID, uuid.New().String(), b.UserID, StatusVals(entity.FuExpNone, "Open", "#fb667e"), nil)
	if err != nil {
		return err
	}
	// add status item - closed
	b.StatusItemClosed, err = b.ItemAdd(ctx, b.StatusEntity.ID, uuid.New().String(), b.UserID, StatusVals(entity.FuExpDone, "Closed", "#66fb99"), nil)
	if err != nil {
		return err
	}
	// add status item - overdue
	b.StatusItemOverDue, err = b.ItemAdd(ctx, b.StatusEntity.ID, uuid.New().String(), b.UserID, StatusVals(entity.FuExpNeg, "OverDue", "#66fb99"), nil)
	if err != nil {
		return err
	}
	fmt.Println("\tCRM:BOOT Status Entity With It's Three Statuses Items Created")

	// add type entity
	b.TypeEntity, err = b.EntityAdd(ctx, uuid.New().String(), schema.SeedTypeEntityName, "Type", entity.CategoryChildUnit, entity.StateTeamLevel, TypeFields())
	if err != nil {
		return err
	}
	// add type item - email
	b.TypeItemEmail, err = b.ItemAdd(ctx, b.TypeEntity.ID, uuid.New().String(), b.UserID, TypeVals(entity.FuExpNone, "Email"), nil)
	if err != nil {
		return err
	}
	// add type item - todo
	b.TypeItemTodo, err = b.ItemAdd(ctx, b.TypeEntity.ID, uuid.New().String(), b.UserID, TypeVals(entity.FuExpNone, "Todo"), nil)
	if err != nil {
		return err
	}

	// add type template - project

	fmt.Println("\tCRM:BOOT Type Entity With It's Three types Items Created")
	return nil
}

func (b *Base) EntityFieldsUpdate(ctx context.Context, entityID string, fields []entity.Field) error {
	input, err := json.Marshal(fields)
	if err != nil {
		return err
	}

	fmt.Printf("\t\tEntity '%s' Updated With Fields\n", entityID)
	return entity.Update(ctx, b.DB, b.AccountID, entityID, string(input), time.Now())
}

func (b *Base) EntityAdd(ctx context.Context, entityID, name, displayName string, category, state int, fields []entity.Field) (entity.Entity, error) {
	ne := entity.NewEntity{
		ID:          entityID,
		AccountID:   b.AccountID,
		TeamID:      b.TeamID,
		Category:    category,
		Name:        name,
		DisplayName: displayName,
		State:       state,
		Fields:      fields,
	}

	e, err := entity.Create(ctx, b.DB, ne, time.Now())
	if err != nil {
		return entity.Entity{}, err
	}

	fmt.Printf("\t\tEntity '%s' Bootstraped\n", e.DisplayName)
	return e, nil
}

func (b *Base) ItemAdd(ctx context.Context, entityID, itemID, userID string, fields map[string]interface{}, source map[string]string) (item.Item, error) {
	name := "System Generated"
	ni := item.NewItem{
		ID:        itemID,
		Name:      &name,
		AccountID: b.AccountID,
		EntityID:  entityID,
		UserID:    &userID,
		Fields:    fields,
		Source:    source,
	}

	it, err := item.Create(ctx, b.DB, ni, time.Now())
	if err != nil {
		return item.Item{}, err
	}

	job.NewJob(b.DB, b.RP, b.FirebaseSDKPath).Stream(stream.NewCreteItemMessage(b.AccountID, userID, entityID, it.ID, ni.Source))

	fmt.Printf("\t\tItem '%s' Bootstraped\n", *it.Name)
	return it, nil
}

func (b *Base) TemplateAdd(ctx context.Context, entityID, itemID, userID string, fields map[string]interface{}, source map[string]string) (item.Item, error) {
	ce, err := entity.Retrieve(ctx, b.AccountID, entityID, b.DB)
	if err != nil {
		return item.Item{}, err
	}

	name := "System Generated"
	ni := item.NewItem{
		ID:        itemID,
		Name:      &name,
		AccountID: b.AccountID,
		EntityID:  entityID,
		UserID:    &userID,
		State:     item.StateBluePrint,
		Fields:    fields,
		Source:    source,
	}

	valueAddedFields := ce.ValueAdd(ni.Fields)
	for _, f := range valueAddedFields {
		if f.IsTitleLayout() {
			s := f.Value.(string)
			ni.Name = &s
		}

		if f.IsDateTime() {
			ni.Fields[f.Key] = fmt.Sprintf("<<%v>>", f.Value)
		}
	}

	it, err := item.Create(ctx, b.DB, ni, time.Now())
	if err != nil {
		return item.Item{}, err
	}

	fmt.Printf("\t\tTemplate '%s' Bootstraped\n", *it.Name)
	return it, nil
}

func (b *Base) FlowAdd(ctx context.Context, flowID, entityID string, name string, mode, condition int, exp string, ftype int) (flow.Flow, error) {
	nf := flow.NewFlow{
		ID:         flowID,
		AccountID:  b.AccountID,
		EntityID:   entityID,
		Mode:       mode,
		Type:       ftype,
		Condition:  condition,
		Expression: exp,
		Name:       name,
	}

	f, err := flow.Create(ctx, b.DB, nf, time.Now())
	if err != nil {
		return flow.Flow{}, err
	}

	fmt.Printf("\t\tFlow '%s' Bootstraped\n", name)
	return f, nil
}

func (b *Base) NodeAdd(ctx context.Context, nodeID, flowID, actorID string, pnodeID string, name string, typ int, exp string, actuals map[string]string, stageID, description string) (node.Node, error) {
	nn := node.NewNode{
		ID:           nodeID,
		AccountID:    b.AccountID,
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

	n, err := node.Create(ctx, b.DB, nn, time.Now())
	if err != nil {
		return node.Node{}, err
	}

	fmt.Printf("\t\t\tNode '%s' Added For Flow %s\n", name, flowID)
	return n, nil
}

func (b *Base) AssociationAdd(ctx context.Context, srcEntityID, dstEntityID string) (string, error) {
	relationshipID, err := relationship.Associate(ctx, b.DB, b.AccountID, srcEntityID, dstEntityID)
	if err != nil {
		return "", err
	}

	fmt.Printf("\t\tAssociation added between entities '%s' and '%s'\n", srcEntityID, dstEntityID)
	return relationshipID, nil
}

func (b *Base) ConnectionAdd(ctx context.Context, relationshipID, entityName, srcEntityID, dstEntityID, srcItemID, dstItemID string, valueAddedFields []entity.Field, action string) error {
	err := connection.Associate(ctx, b.DB, b.AccountID, b.UserID, relationshipID, entityName, srcEntityID, dstEntityID, srcItemID, dstItemID, valueAddedFields, action)
	if err != nil {
		return err
	}
	fmt.Printf("\t\t\tConnection added between items '%s' and '%s' for the relationship '%s'\n", srcItemID, dstItemID, relationshipID)
	return nil
}

func (b *Base) LayoutAdd(ctx context.Context, name, entityID string, fields map[string]string) error {
	nl := layout.NewLayout{}
	nl.Name = name
	nl.AccountID = b.AccountID
	nl.EntityID = entityID
	nl.Fields = fields

	_, err := layout.Create(ctx, b.DB, nl, time.Now())
	return err
}
