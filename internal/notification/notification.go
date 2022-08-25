package notification

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/visitor"
)

type NotificationType int

const (
	TypeReminder NotificationType = iota
	TypeAssigned
	TypeCreated
	TypeUpdated
	TypeMemberInvitation
	TypeWelcome
	TypeVisitorInvitation
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
	return emailNotif.Send(ctx, TypeMemberInvitation, db)
}

func VisitorInvitation(accountID, visitorID, body string, db *sqlx.DB, rp *redis.Pool) error {
	ctx := context.Background()

	a, err := account.Retrieve(ctx, db, accountID)
	if err != nil {
		log.Println("***>***> VisitorInvitation: unexpected/unhandled error occurred when retriving account. error:", err)
		return err
	}

	v, err := visitor.Retrieve(ctx, accountID, visitorID, db)
	if err != nil {
		log.Println("***>***> VisitorInvitation: unexpected/unhandled error occurred when retriving visitor. error:", err)
		return err
	}

	usrName := util.NameInEmail(v.Email)
	magicLink, err := auth.CreateVisitorMagicLink(accountID, usrName, v.Email, visitorID, v.Token, rp)
	if err != nil {
		log.Println("***>***> VisitorInvitation: unexpected/unhandled error occurred when creating the magic link. error:", err)
		return err
	}

	emailNotif := EmailNotification{
		To:          []interface{}{v.Email},
		Subject:     fmt.Sprintf("Invitation to visit %s account", a.Name),
		Name:        usrName,
		Requester:   fmt.Sprintf("Admin of %s account", a.Name),
		AccountName: a.Name,
		MagicLink:   magicLink,
		Body:        body,
	}
	return emailNotif.Send(ctx, TypeVisitorInvitation, db)
}

func OnAnItemLevelEvent(ctx context.Context, usrID string, entityCategory int, entityDisName, accountID, teamID, entityID, itemID string, itemCreatorID *string, valueAddedFields []entity.Field, dirtyFields map[string]interface{}, baseIds []string, notificationType NotificationType, db *sqlx.DB, firebaseSDKPath string) (*item.Item, error) {
	appNotif := appNotificationBuilder(ctx, accountID, teamID, usrID, entityID, itemID, itemCreatorID, valueAddedFields, dirtyFields, baseIds, db)

	switch notificationType {
	case TypeReminder:
		if val, exist := appNotif.DirtyFields[entity.WhoDueBy]; exist {
			appNotif.Subject = fmt.Sprintf("%s is due on %s", util.UpperSinglarize(entityDisName), val)
		}
		appNotif.Body = fmt.Sprintf("%s", appNotif.Title)
	case TypeCreated:
		switch entityCategory {
		case entity.CategoryEmail:
			appNotif.Subject = fmt.Sprintf("A new e-mail has been sent/received")
			appNotif.Body = fmt.Sprintf("%s", appNotif.Title)
		case entity.CategoryUsers:
			appNotif.Subject = fmt.Sprintf("A new member added to your account")
			appNotif.Body = fmt.Sprintf("New member %s added to your account", appNotif.Title)
		default:
			appNotif.Subject = fmt.Sprintf("A new record created in %s", util.LowerPluralize(entityDisName))
			appNotif.Body = fmt.Sprintf("%s", appNotif.Title)
		}

	case TypeUpdated:
		appNotif.Subject = fmt.Sprintf("A record in %s is updated", util.LowerPluralize(entityDisName))
		appNotif.Body = fmt.Sprintf("%s", appNotif.Title)
		if val, exist := appNotif.DirtyFields[entity.WhoAssignee]; exist {
			appNotif.Body = fmt.Sprintf("%s `%s` has been updated with assignee/s %s", util.UpperSinglarize(entityDisName), appNotif.Title, val)
		} else if val, exist := appNotif.DirtyFields[entity.WhoDueBy]; exist {
			appNotif.Body = fmt.Sprintf("%s `%s` due date has been modified %s", util.UpperSinglarize(entityDisName), appNotif.Title, val)
		} else if val, exist := appNotif.DirtyFields["modified_fields"]; exist {
			appNotif.Body = fmt.Sprintf("%s `%s` has been modified with the following fields %s", util.UpperSinglarize(entityDisName), appNotif.Title, val)
		}
	}

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
		Name:      strings.Title(assignee.Name),
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
