package engine

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/visitor"
)

func (eng *Engine) executeInvite(ctx context.Context, n node.Node, db *sqlx.DB, rp *redis.Pool) error {
	e, err := entity.Retrieve(ctx, n.AccountID, n.Meta.EntityID, db)
	if err != nil {
		return err
	}

	i, err := item.Retrieve(ctx, n.Meta.EntityID, n.Meta.ItemID, db)
	if err != nil {
		return err
	}

	if e.EmailField() != nil {
		email := i.Fields()[e.EmailField().Key]
		if email != nil {

			token, _ := auth.GenerateRandomToken(32)
			if err != nil {
				return err
			}

			visitorInvitationFields, err := valueAdd(ctx, db, n.AccountID, n.ActorID, n.ActualsItemID())
			if err != nil {
				return err
			}
			namedFieldsObj := entity.NamedFieldsObjMap(visitorInvitationFields)
			body := namedFieldsObj["body"].Value.(string)
			body = eng.RunExpRenderer(ctx, db, n.AccountID, body, n.VariablesMap())

			nv := visitor.NewVisitor{
				AccountID: e.AccountID,
				TeamID:    e.TeamID,
				EntityID:  e.ID,
				ItemID:    i.ID,
				Email:     email.(string),
				Token:     token,
			}

			v, err := visitor.Create(ctx, db, nv, time.Now())
			if err != nil {
				return err
			}

			err = eng.Job.AddVisitor(v.AccountID, v.VistitorID, body, db, rp)
			return err
		}
	}

	//a user should be either USER or ADMIN/SUPER-ADMIN can't be both
	//add the user-id/account-id/team-id/entity-id/item-id to the access-list table.
	//send it to user via email
	return nil
}
