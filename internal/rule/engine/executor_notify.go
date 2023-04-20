package engine

import (
	"context"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/voip"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/token"
)

func (eng *Engine) executeNotify(ctx context.Context, n node.Node, db *sqlx.DB, sdb *database.SecDB) error {
	notifyFields, err := valueAdd(ctx, db, sdb, n.AccountID, n.ActorID, n.ActualsItemID())
	if err != nil {
		return err
	}

	var ownerEntityID string
	for i := 0; i < len(notifyFields); i++ {
		ntf := notifyFields[i]
		if ntf.Who == entity.WhoAssignee {
			ownerEntityID = ntf.RefID
			break
		}
	}

	var notifyEntityItem entity.NotifyEntity
	err = entity.ParseFixedEntity(notifyFields, &notifyEntityItem)
	if err != nil {
		return err
	}

	if ownerEntityID != "" && len(notifyEntityItem.Owner) > 0 {
		userItem, err := entity.RetriveUserItem(ctx, n.AccountID, ownerEntityID, notifyEntityItem.Owner[0], db, sdb)
		if err != nil {
			return err
		}
		tkn, err := token.Retrieve(ctx, db, n.AccountID)
		if err != nil {
			return err
		}

		twilioSID := util.ExpvarGet("twilio_sid")
		twilioToken := util.ExpvarGet("twilio_token")
		source := n.VarStrMap()
		for baseEntityID, baseItemIDs := range source {
			if len(baseItemIDs) > 0 {
				//dev
				if userItem.Phone == "" {
					userItem.Phone = "+14083485853"
				}
				err = voip.MakeCall(twilioSID, twilioToken, n.AccountID, tkn.Token, baseEntityID, baseItemIDs[0], userItem.Phone)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
