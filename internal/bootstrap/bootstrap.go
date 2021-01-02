package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

const (
	OwnerEntity = "owners"
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

func BootstrapUserEntity(ctx context.Context, db *sqlx.DB, currentUser *user.User, accountID, teamID string) error {
	fields, itemVals := ownerFields(currentUser.ID, *currentUser.Name, *currentUser.Avatar, currentUser.Email)
	// add entity - owners
	ue, err := EntityAdd(ctx, db, accountID, teamID, uuid.New().String(), OwnerEntity, "Owners", entity.CategoryUsers, fields)
	if err != nil {
		return err
	}
	// add owner item
	_, err = ItemAdd(ctx, db, accountID, ue.ID, uuid.New().String(), itemVals)
	if err != nil {
		return err
	}
	return nil
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

	fmt.Printf("Entity '%s' Bootstraped\n", e.DisplayName)
	return e, nil
}

func ItemAdd(ctx context.Context, db *sqlx.DB, accountID, entityID, itemID string, fields map[string]interface{}) (item.Item, error) {
	ni := item.NewItem{
		ID:        itemID,
		AccountID: accountID,
		EntityID:  entityID,
		Fields:    fields,
	}

	i, err := item.Create(ctx, db, ni, time.Now())
	if err != nil {
		return item.Item{}, err
	}

	job.OnFieldCreate(accountID, entityID, ni.ID, ni.Fields, db)

	fmt.Printf("%s - Item Added\n", i.ID)
	return i, nil
}

func FlowAdd(ctx context.Context, db *sqlx.DB, accountID, flowID, entityID string, name string, mode, condition int) (flow.Flow, error) {
	nf := flow.NewFlow{
		ID:         flowID,
		AccountID:  accountID,
		EntityID:   entityID,
		Mode:       mode,
		Type:       flow.FlowTypeFieldUpdate,
		Condition:  condition,
		Expression: `{{` + entityID + `.uuid-00-fname}} eq {Vijay} && {{` + entityID + `.uuid-00-nps-score}} gt {98}`,
		Name:       name,
	}

	f, err := flow.Create(ctx, db, nf, time.Now())
	if err != nil {
		return flow.Flow{}, err
	}

	fmt.Printf("Flow '%s' Bootstraped\n", name)
	return f, nil
}

func NodeAdd(ctx context.Context, db *sqlx.DB, accountID, nodeID, flowID, actorID string, pnodeID string, name string, typ int, exp string, actuals map[string]string) (node.Node, error) {
	nn := node.NewNode{
		ID:           nodeID,
		AccountID:    accountID,
		FlowID:       flowID,
		ActorID:      actorID,
		ParentNodeID: pnodeID,
		Name:         name,
		Type:         typ,
		Expression:   exp,
		Actuals:      actuals,
	}

	n, err := node.Create(ctx, db, nn, time.Now())
	if err != nil {
		return node.Node{}, err
	}

	fmt.Printf("Node '%s' Added For Flow %s\n", name, flowID)
	return n, nil
}

func AssociationAdd(ctx context.Context, db *sqlx.DB, accountID, srcEntityID, dstEntityID string) (string, error) {
	relationshipID, err := entity.Associate(ctx, db, accountID, srcEntityID, dstEntityID)
	if err != nil {
		return "", err
	}

	fmt.Printf("Association '%s' Added\n", relationshipID)
	return relationshipID, nil
}

func ConnectionAdd(ctx context.Context, db *sqlx.DB, accountID, relationshipID, srcItemID, dstItemID string) error {
	err := item.Associate(ctx, db, accountID, relationshipID, srcItemID, dstItemID)
	if err != nil {
		return err
	}
	fmt.Printf("Connection '%s' Added\n", relationshipID)
	return nil
}
