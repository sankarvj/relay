package notification

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

type FirebaseNotification struct {
	AccountID string
	TeamID    string
	EntityID  string
	ItemID    string
	MemberID  string
	Subject   string
	Body      string
}

func (fbNotif FirebaseNotification) Send(ctx context.Context, notifType NotificationType, db *sqlx.DB) error {
	log.Println("********************>>>>>> fbNotif.MemberID ", fbNotif.MemberID)
	if fbNotif.MemberID == "" {
		return nil
	}

	ownerEntity, err := entity.RetrieveFixedEntity(ctx, db, fbNotif.AccountID, "", entity.FixedEntityOwner)
	if err != nil {
		log.Println("********************>>>>>> fbNotif.AccountID ", fbNotif.AccountID)
		log.Println("********************>>>>>> fbNotif.err ", err)
		return err
	}
	log.Println("********************>>>>>> ownerEntity ", ownerEntity)
	var userEntityItem entity.UserEntity
	valueAddedFields, _, err := entity.RetrieveFixedItem(ctx, ownerEntity.AccountID, ownerEntity.ID, fbNotif.MemberID, db)
	if err != nil {
		return err
	}
	err = entity.ParseFixedEntity(valueAddedFields, &userEntityItem)
	if err != nil {
		return err
	}
	log.Println("********************>>>>>> userEntityItem userEntityItem.UserID ", userEntityItem.UserID)

	client, err := RetrieveClient(ctx, fbNotif.AccountID, userEntityItem.UserID, db)
	if err != nil {
		return err
	}
	log.Printf("********************>>>>>> userEntityItem client %+v", client)

	return nil
}
