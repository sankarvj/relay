package notification

import (
	"context"
	"fmt"
	"log"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
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

func OnAnItemLevelEvent(ctx context.Context, usrID, entityName string, accountID, teamID, entityID, itemID string, itemCreatorID *string, valueAddedFields []entity.Field, dirtyFields map[string]interface{}, notificationType NotificationType, db *sqlx.DB, firebaseSDKPath string) (*item.Item, error) {
	appNotif := appNotificationBuilder(ctx, accountID, teamID, usrID, entityID, itemID, itemCreatorID, valueAddedFields, dirtyFields, db)

	switch notificationType {
	case TypeReminder:
		if val, exist := appNotif.DirtyFields[entity.WhoDueBy]; exist {
			appNotif.Subject = fmt.Sprintf("Your `%s` is due on %s", entityName, val)
		}
		appNotif.Body = fmt.Sprintf("%s...", appNotif.Title)
	case TypeCreated:
		appNotif.Subject = fmt.Sprintf("An item created `%s` module", entityName)
		appNotif.Body = fmt.Sprintf("A new %s `%s` added", entityName, appNotif.Title)
	case TypeUpdated:
		appNotif.Subject = fmt.Sprintf("%s `%s` is updated", entityName, appNotif.Title)
		appNotif.Body = fmt.Sprintf("%s...", appNotif.Title)
		if val, exist := appNotif.DirtyFields[entity.WhoAssignee]; exist {
			appNotif.Body = fmt.Sprintf("%s `%s` has been updated with assignee/s %s", entityName, appNotif.Title, val)
		} else if val, exist := appNotif.DirtyFields[entity.WhoDueBy]; exist {
			appNotif.Body = fmt.Sprintf("%s `%s` due date has been modified %s", entityName, appNotif.Title, val)
		} else if val, exist := appNotif.DirtyFields["modified_fields"]; exist {
			appNotif.Body = fmt.Sprintf("%s `%s` has been modified with the following fields %s", entityName, appNotif.Title, val)
		}
	}

	log.Println("appNotif.Assignees ", appNotif.Assignees)

	//Send email/firebase notification to assignees/followers/creators
	for _, assignee := range appNotif.Assignees {
		sendEmailAndFBNotification(ctx, appNotif, assignee, notificationType, db, firebaseSDKPath)
	}

	if itemCreatorID != nil && *itemCreatorID != usrID {
		for _, follower := range appNotif.Followers {
			sendEmailAndFBNotification(ctx, appNotif, follower, notificationType, db, firebaseSDKPath)
		}
	}

	return appNotif.Send(ctx, notificationType, db)
}

func sendEmailAndFBNotification(ctx context.Context, appNotif AppNotification, assignee entity.UserEntity, notificationType NotificationType, db *sqlx.DB, firebaseSDKPath string) (error, error) {

	emailNotif := EmailNotification{
		To:        []interface{}{assignee.Email},
		Subject:   appNotif.Subject,
		Body:      appNotif.Body,
		MagicLink: auth.SimpleLink(appNotif.AccountID, appNotif.TeamID, appNotif.EntityID, appNotif.ItemID),
	}
	err1 := emailNotif.Send(ctx, notificationType, db)
	if err1 != nil {
		log.Println("***>***> emailNotif.Send. error:", err1)
	}

	fbNotif := FirebaseNotification{
		AccountID: appNotif.AccountID,
		UserID:    assignee.UserID,
		Subject:   appNotif.Subject,
		Body:      appNotif.Body,
		SDKPath:   firebaseSDKPath,
	}
	err2 := fbNotif.Send(ctx, notificationType, db)
	if err2 != nil {
		log.Println("***>***> fbNotif.Send. error:", err2)
	}
	return err1, err2
}
