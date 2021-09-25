package notification

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type GCMNotification struct {
}

func (gcmNotif GCMNotification) Send(ctx context.Context, notifType NotificationType, db *sqlx.DB) error {
	return nil
}
