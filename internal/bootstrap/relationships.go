package bootstrap

import (
	"context"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func addAssociations(ctx context.Context, db *sqlx.DB, accountID, teamID string, conEid, comEid, deEid, tickEid string, conID, comID, dealID, ticketID string, emailKey string, emailsEntity entity.Entity) error {
	//contact company association
	associationID, err := AssociationAdd(ctx, db, accountID, conEid, comEid)
	if err != nil {
		return err
	}
	err = ConnectionAdd(ctx, db, accountID, associationID, conID, comID)
	if err != nil {
		return err
	}

	//ticket deal association
	tdaID, err := AssociationAdd(ctx, db, accountID, tickEid, deEid)
	if err != nil {
		return err
	}
	err = ConnectionAdd(ctx, db, accountID, tdaID, ticketID, dealID)
	if err != nil {
		return err
	}

	//deal email association
	_, err = AssociationAdd(ctx, db, accountID, deEid, emailsEntity.ID)
	if err != nil {
		return err
	}

	return nil
}
