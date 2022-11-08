package engine

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	rg "github.com/redislabs/redisgraph-go"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"gitlab.com/vjsideprojects/relay/internal/visitor"
)

func (eng *Engine) executeInvite(ctx context.Context, n node.Node, db *sqlx.DB, sdb *database.SecDB) error {

	e, err := entity.Retrieve(ctx, n.AccountID, n.Meta.EntityID, db, sdb)
	if err != nil {
		return err
	}

	i, err := item.Retrieve(ctx, n.AccountID, n.Meta.EntityID, n.Meta.ItemID, db)
	if err != nil {
		return err
	}

	visitorInvitationFields, err := valueAdd(ctx, db, sdb, n.AccountID, n.ActorID, n.ActualsItemID())
	if err != nil {
		return err
	}

	namedFieldsObj := entity.NameMap(visitorInvitationFields)
	body := namedFieldsObj["body"].Value.(string)
	role := namedFieldsObj["role"].Value.([]interface{})
	body = eng.RunExpRenderer(ctx, db, sdb, n.AccountID, body, n.VariablesMap())
	email := namedFieldsObj["email"].Value.(string)
	email = eng.RunExpRenderer(ctx, db, sdb, n.AccountID, email, n.VariablesMap())

	emails := strings.Split(email, ",")
	for _, email := range emails {
		if email != "" && util.IsValidEmail(email) {
			token, _ := auth.GenerateRandomToken(32)
			if err != nil {
				return err
			}

			nv := visitor.NewVisitor{
				AccountID: e.AccountID,
				TeamID:    e.TeamID,
				EntityID:  e.ID,
				ItemID:    i.ID,
				Name:      util.NameInEmail(email),
				Email:     email,
				Token:     token,
			}

			if len(role) > 0 {
				if role[0] == auth.RoleAdmin || role[0] == auth.RoleMember || role[0] == auth.RoleUser {
					return inviteMember(ctx, nv, role[0].(string), body, eng, db, sdb)
				} else if role[0] == auth.RoleVisitor {
					return inviteVisitor(ctx, nv, body, eng, db, sdb)
				}
			}
		}
	}

	return nil
}

//a user should be either USER or ADMIN/SUPER-ADMIN can't be both
//add the user-id/account-id/team-id/entity-id/item-id to the access-list table.
//send it to user via email
func inviteVisitor(ctx context.Context, nv visitor.NewVisitor, body string, eng *Engine, db *sqlx.DB, sdb *database.SecDB) error {
	v, err := visitor.Create(ctx, db, nv, time.Now())
	if err != nil {
		return err
	}

	err = eng.Job.AddVisitor(v.AccountID, v.VistitorID, body, db, sdb)
	return err
}

func inviteMember(ctx context.Context, nv visitor.NewVisitor, role string, body string, eng *Engine, db *sqlx.DB, sdb *database.SecDB) error {
	//adding access for this member..
	if role == auth.RoleUser {
		_, err := visitor.Create(ctx, db, nv, time.Now())
		if err != nil {
			return err
		}
	}

	e, err := entity.RetrieveByName(ctx, nv.AccountID, entity.FixedEntityOwner, db)
	if err != nil {
		return err
	}

	namedKeys := entity.NameMap(e.EasyFields())

	items, err := checkIfMemberAlreadyExist(ctx, nv.AccountID, e.ID, nv.Email, namedKeys, db, sdb)
	if err != nil {
		return err
	}

	var memberID string
	var memberName string
	var memberEmail string
	if len(items) > 0 {
		for _, i := range items {
			itemFields := i.Fields()
			teamIds := itemFields[namedKeys["team_ids"].Key].([]interface{})
			teamIds = append(teamIds, nv.TeamID)
			itemFields[namedKeys["team_ids"].Key] = teamIds
			_, err := item.UpdateFields(ctx, db, nv.AccountID, i.EntityID, i.ID, itemFields)
			if err != nil {
				return err
			}
			memberName = itemFields[namedKeys["name"].Key].(string)
			memberEmail = itemFields[namedKeys["email"].Key].(string)
		}
	} else {
		ni := item.NewItem{
			ID:        uuid.New().String(),
			AccountID: e.AccountID,
			EntityID:  e.ID,
			UserID:    util.String(user.UUID_SYSTEM_USER),
			Fields:    recreateFields(nv.Email, role, nv.TeamID, namedKeys),
		}

		it, err := item.Create(ctx, db, ni, time.Now())
		if err != nil {
			return err
		}
		memberID = it.ID
		itemFields := it.Fields()
		memberName = itemFields[namedKeys["name"].Key].(string)
		memberEmail = itemFields[namedKeys["email"].Key].(string)
	}

	err = eng.Job.AddMember(nv.AccountID, memberID, memberName, memberEmail, body, db, sdb)
	return err
}

func recreateFields(email, role, associatedTeamID string, namedKeys map[string]entity.Field) map[string]interface{} {
	itemFields := make(map[string]interface{}, 0)
	itemFields[namedKeys["name"].Key] = util.NameInEmail(email)
	itemFields[namedKeys["user_id"].Key] = ""
	itemFields[namedKeys["email"].Key] = email
	itemFields[namedKeys["avatar"].Key] = ""
	itemFields[namedKeys["team_ids"].Key] = []interface{}{associatedTeamID}
	itemFields[namedKeys["role"].Key] = []interface{}{role}
	return itemFields
}

func checkIfMemberAlreadyExist(ctx context.Context, accountID, entityID, email string, namedKeys map[string]entity.Field, db *sqlx.DB, sdb *database.SecDB) ([]item.Item, error) {
	conditionFields := make([]graphdb.Field, 0)
	emailField := namedKeys["email"]
	gf := graphdb.Field{
		Expression: "=",
		Key:        emailField.Key,
		DataType:   graphdb.DType(emailField.DataType),
		Value:      email,
	}
	conditionFields = append(conditionFields, gf)

	gSegment := graphdb.BuildGNode(accountID, entityID, false, nil).MakeBaseGNode("", conditionFields)
	result, err := graphdb.GetResult(sdb.GraphPool(), gSegment, 0, "", "")
	if err != nil {
		return nil, err
	}

	items, err := itemsResp(ctx, db, accountID, result)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func itemsResp(ctx context.Context, db *sqlx.DB, accountID string, result *rg.QueryResult) ([]item.Item, error) {
	itemIDs := util.ParseGraphResult(result)
	items, err := item.BulkRetrieveItems(ctx, accountID, itemIDs, db)
	if err != nil {
		return []item.Item{}, err
	}

	return items, nil
}
