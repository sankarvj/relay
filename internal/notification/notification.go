package notification

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
)

type NotificationType int

const (
	TypeReminder NotificationType = iota
	TypeAssigned
	TypeCreated
	TypeUpdated
	TypeInvitation
	TypeWelcome
)

type Notification interface {
	Send(ctx context.Context, notifType NotificationType, db *sqlx.DB) error
}

func WelcomeInvitation(draftID, accountName, requester, usrName, usrEmail string, db *sqlx.DB, rp *redis.Pool) error {
	ctx := context.Background()
	magicLink, err := auth.CreateMagicLaunchLink(draftID, accountName, usrEmail, rp)
	if err != nil {
		log.Println("***>***> WelcomeInvitation: unexpected/unhandled error occurred when creating the magic link. error:", err)
		return err
	}

	emailNotif := EmailNotification{
		To:          []interface{}{usrEmail},
		Subject:     fmt.Sprintf("Welcome %s! Get started with workbaseONE", requester),
		Name:        requester,
		Requester:   requester,
		AccountName: accountName,
		MagicLink:   magicLink,
	}
	return emailNotif.Send(ctx, TypeWelcome, db)
}

func JoinInvitation(accountID, accountName, requester, usrName, usrEmail string, memberID string, db *sqlx.DB, rp *redis.Pool) error {
	ctx := context.Background()
	magicLink, err := auth.CreateMagicLink(accountID, usrName, usrEmail, memberID, rp)
	if err != nil {
		log.Println("***>***> JoinInvitation: unexpected/unhandled error occurred when creating the magic link. error:", err)
		return err
	}

	emailNotif := EmailNotification{
		To:          []interface{}{usrEmail},
		Subject:     fmt.Sprintf("Invitation to join %s account", accountName),
		Name:        usrName,
		Requester:   requester,
		AccountName: accountName,
		MagicLink:   magicLink,
	}
	return emailNotif.Send(ctx, TypeInvitation, db)
}

func OnAnItemLevelEvent(ctx context.Context, usrID, entityName string, accountID, teamID, entityID, itemID string, itemCreatorID *string, valueAddedFields []entity.Field, dirtyFields map[string]interface{}, notificationType NotificationType, db *sqlx.DB, firebaseSDKPath string) error {
	appNotif := appNotificationBuilder(ctx, accountID, teamID, usrID, entityID, itemID, itemCreatorID, valueAddedFields, dirtyFields, db)

	switch notificationType {
	case TypeReminder:
		appNotif.Subject = fmt.Sprintf("Your `%s` is due on %s", entityName, appNotif.DueBy)
		appNotif.Body = fmt.Sprintf("%s...", appNotif.Title)
	case TypeCreated:
		appNotif.Subject = fmt.Sprintf("A new `%s` created", entityName)
		appNotif.Body = fmt.Sprintf("%s...", appNotif.Title)
	case TypeUpdated:
		appNotif.Subject = fmt.Sprintf("A `%s` item has been updated", entityName)
		appNotif.Body = fmt.Sprintf("%s...", appNotif.Title)
		if appNotif.Names != "" {
			appNotif.Body = fmt.Sprintf("Module `%s` has been updated with assignee/s %s", appNotif.Title, appNotif.Names)
		} else if appNotif.DueBy != "" {
			appNotif.Body = fmt.Sprintf("Module `%s` due date has been modified %s", appNotif.Title, appNotif.DueBy)
		} else {
			appNotif.Body = fmt.Sprintf("Module `%s` has been updated following fields %s", appNotif.Title, strings.Join(appNotif.ModifiedFields, ","))
		}
	}

	//Send email/firebase notification to assignees/followers/creators
	for _, assignee := range appNotif.Assignees {
		emailNotif := EmailNotification{
			To:      []interface{}{assignee.Email},
			Subject: appNotif.Subject,
			Body:    appNotif.Body,
		}
		emailNotif.Send(ctx, notificationType, db)

		fbNotif := FirebaseNotification{
			AccountID: appNotif.AccountID,
			UserID:    assignee.UserID,
			Subject:   appNotif.Subject,
			Body:      appNotif.Body,
			SDKPath:   firebaseSDKPath,
		}
		fbNotif.Send(ctx, notificationType, db)
	}

	return appNotif.Send(ctx, notificationType, db)
}
