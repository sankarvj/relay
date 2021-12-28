package notification

import (
	"context"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

type AppNotification struct {
	AccountID string
	TeamID    string
	EntityID  string
	ItemID    string
	UserID    string
	Subject   string
	Body      string
}

func (appNotif AppNotification) Send(ctx context.Context, notifType NotificationType, db *sqlx.DB) error {

	notificationItem := entity.NotificationEntityItem{
		AccountID: appNotif.AccountID,
		EntityID:  appNotif.EntityID,
		ItemID:    appNotif.ItemID,
		Subject:   appNotif.Subject,
		Body:      appNotif.Body,
		Type:      0, //TODO change this or remove this
		CreatedAt: "some time",
	}

	_, err := entity.SaveFixedEntityItem(ctx, appNotif.AccountID, appNotif.TeamID, appNotif.UserID, entity.FixedEntityNotification, "Notification", "", "", util.ConvertInterfaceToMap(notificationItem), db)
	if err != nil {
		return err
	}
	return nil
}
