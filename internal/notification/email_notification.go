package notification

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type EmailNotification struct {
	Subject string
	Body    string
}

func (emNotif EmailNotification) Send(ctx context.Context, notifType NotificationType, db *sqlx.DB) error {
	// send email using aws SES if email notification enabled.
	// from: no-reply@baserelay.com
	// to: <individual user who got assigned> , <updates to the user if already assigned>, <@mention on the notes/conversations>
	return nil
}
