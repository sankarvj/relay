package notification

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

type NotificationType int

const (
	TypeReminder NotificationType = iota
	TypeAssigned
	TypeCreated
	TypeUpdated
)

type Notification interface {
	Send(ctx context.Context, notifType NotificationType, db *sqlx.DB) error
}

func UserInvitation(ctx context.Context) error {
	return nil
}

func ItemUpdates(ctx context.Context, name string, accountID, teamID, entityID, itemID string, valueAddedFields []entity.Field, notificationType NotificationType, db *sqlx.DB) error {
	var subject string
	var body string
	var formettedTime string
	for _, f := range valueAddedFields {
		if f.IsTitleLayout() {
			body = f.Value.(string)
		}

		if f.Who == entity.WhoDueBy && f.DataType == entity.TypeDateTime && f.Value != nil {
			when, _ := util.ParseTime(f.Value.(string))
			formettedTime = util.FormatTimeGo(when)
		}
	}

	switch notificationType {
	case TypeReminder:
		subject = fmt.Sprintf("your %s is due on %s", name, formettedTime)
	case TypeAssigned:
	case TypeCreated:
	case TypeUpdated:
	}

	notif := AppNotification{
		AccountID: accountID,
		TeamID:    teamID,
		EntityID:  entityID,
		ItemID:    itemID,
		Subject:   subject,
		Body:      body,
	}

	return notif.Send(ctx, notificationType, db)
}