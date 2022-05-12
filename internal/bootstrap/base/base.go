package base

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/connection"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/layout"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

const (
	UUIDHolder = "00000000-0000-0000-0000-000000000000"
)

type Base struct {
	AccountID string
	TeamID    string
	UserID    string
	DB        *sqlx.DB
	RP        *redis.Pool
}

func NewBase(accountID, teamID, userID string, db *sqlx.DB, rp *redis.Pool) *Base {
	return &Base{
		AccountID: accountID,
		TeamID:    teamID,
		UserID:    userID,
		DB:        db,
		RP:        rp,
	}
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

func (b *Base) ItemAdd(ctx context.Context, entityID, itemID, userID string, fields map[string]interface{}) (item.Item, error) {
	return b.ItemAddGenie(ctx, entityID, itemID, userID, UUIDHolder, fields, nil)
}

func (b *Base) ItemAddGenie(ctx context.Context, entityID, itemID, userID, genieID string, fields map[string]interface{}, source map[string]string) (item.Item, error) {
	name := "System Generated"
	ni := item.NewItem{
		ID:        itemID,
		Name:      &name,
		AccountID: b.AccountID,
		EntityID:  entityID,
		UserID:    &userID,
		GenieID:   &genieID,
		Fields:    fields,
		Source:    source,
	}

	it, err := item.Create(ctx, b.DB, ni, time.Now())
	if err != nil {
		return item.Item{}, err
	}

	job.NewJob(b.DB, b.RP).Stream(stream.NewCreteItemMessage(b.AccountID, userID, entityID, it.ID, ni.Source))

	fmt.Printf("\t\tItem '%s' Bootstraped\n", *it.Name)
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
