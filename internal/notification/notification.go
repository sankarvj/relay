package notification

import (
	"context"
	"fmt"
	"log"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/draft"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/visitor"
)

type NotificationType int

const (
	TypeReminder               NotificationType = 0
	TypeAssigned               NotificationType = 1
	TypeCreated                NotificationType = 2
	TypeUpdated                NotificationType = 3
	TypeMemberInvitation       NotificationType = 4
	TypeWelcome                NotificationType = 5
	TypeVisitorInvitation      NotificationType = 6
	TypeEmailConversationAdded NotificationType = 7
	TypeChatConversationAdded  NotificationType = 8
	TypeMemberAdded            NotificationType = 9
)

type Notification interface {
	Send(ctx context.Context, notifType NotificationType, db *sqlx.DB) error
}

func WelcomeInvitation(draftID string, teams []string, accountName, requester, usrName, usrEmail string, db *sqlx.DB, rp *redis.Pool) error {
	ctx := context.Background()

	workBaseDomain := "workbaseone.com"
	if len(teams) == 1 {
		workBaseDomain = draft.TeamDomainMap[teams[0]]
	}

	magicLink, err := auth.CreateMagicLaunchLink(workBaseDomain, draftID, accountName, usrEmail, rp)
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

func JoinInvitation(accountID, accountName string, teams []string, requester, usrName, usrEmail string, memberID string, db *sqlx.DB, rp *redis.Pool) error {
	ctx := context.Background()
	workBaseDomain := "workbaseone.com"
	if len(teams) == 1 {
		workBaseDomain = draft.TeamDomainMap[teams[0]]
	}

	magicLink, err := auth.CreateMagicLink(workBaseDomain, accountID, usrName, usrEmail, memberID, rp)
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

func OnAnItemLevelEvent(ctx context.Context, usrID string, entityCategory int, entityDisName, accountID, teamID, entityID, itemID string, itemCreatorID *string, valueAddedFields []entity.Field, dirtyFields map[string]interface{}, source map[string][]string, notificationType NotificationType, db *sqlx.DB, firebaseSDKPath string) (*item.Item, error) {
	appNotif := appNotificationBuilder(ctx, accountID, teamID, usrID, entityID, itemID, itemCreatorID, valueAddedFields, dirtyFields, source, db)

	switch notificationType {
	case TypeReminder:
		if val, exist := appNotif.DirtyFields[entity.WhoDueBy]; exist {
			appNotif.Subject = fmt.Sprintf("%s is due on %s", util.UpperSinglarize(entityDisName), val)
		}
		appNotif.Body = fmt.Sprintf("%s", appNotif.Title)
	case TypeCreated:
		switch entityCategory {
		case entity.CategoryEmail:
			appNotif.Subject = fmt.Sprintf("An e-mail has been sent/received")
			// enriching subject with base elements
			if appNotif.BaseEntityName != "" {
				appNotif.Subject = fmt.Sprintf("%s for %s", appNotif.Subject, util.LowerSinglarize(appNotif.BaseEntityName))
				if appNotif.BaseItemName != "" {
					appNotif.Subject = fmt.Sprintf("%s `%s`", appNotif.Subject, appNotif.BaseItemName)
				}
			}
			appNotif.Body = fmt.Sprintf("%s", appNotif.Title)
		case entity.CategoryUsers:
			notificationType = TypeMemberAdded //changing notification type in the mid-way
			// adding all members as followers
			appNotif.AddMoreFollowers(ctx, accountID, db)
			appNotif.Subject = fmt.Sprintf("A new member added to your account")
			appNotif.Body = fmt.Sprintf("New member %s added to your account", appNotif.Title)
		default:
			if len(source) == 0 { // add all members for the main module addition...
				appNotif.AddMoreFollowers(ctx, accountID, db)
			}

			appNotif.Subject = fmt.Sprintf("A new %s created", util.LowerSinglarize(entityDisName))
			// enriching subject with base elements
			if appNotif.BaseEntityName != "" {
				appNotif.Subject = fmt.Sprintf("%s in %s", appNotif.Subject, util.LowerSinglarize(appNotif.BaseEntityName))
				if appNotif.BaseItemName != "" {
					appNotif.Subject = fmt.Sprintf("%s `%s`", appNotif.Subject, appNotif.BaseItemName)
				}
			}
			appNotif.Body = fmt.Sprintf("%s", appNotif.Title)
		}

	case TypeUpdated:
		appNotif.Subject = fmt.Sprintf("A record in %s is updated", util.LowerPluralize(entityDisName))
		appNotif.Body = fmt.Sprintf("%s", appNotif.Title)
		switch entityCategory {
		case entity.CategoryUsers:
			appNotif.Subject = "Invited member logged in"
			appNotif.Body = fmt.Sprintf("%s `%s` has been logged in for the first time", "Member", appNotif.Title)
		default:
			if val, exist := appNotif.DirtyFields[entity.WhoAssignee]; exist {
				appNotif.Body = fmt.Sprintf("%s `%s` has been updated with assignee/s %s", util.UpperSinglarize(entityDisName), appNotif.Title, val)
			} else if val, exist := appNotif.DirtyFields[entity.WhoDueBy]; exist {
				appNotif.Body = fmt.Sprintf("%s `%s` due date has been modified %s", util.UpperSinglarize(entityDisName), appNotif.Title, val)
			} else if val, exist := appNotif.DirtyFields["modified_fields"]; exist {
				appNotif.Body = fmt.Sprintf("%s `%s` has been modified with the following fields %s", util.UpperSinglarize(entityDisName), appNotif.Title, val)
			}
		}

	case TypeChatConversationAdded:
		if val, exist := appNotif.DirtyFields[entity.WhoMessage]; exist {
			appNotif.Body = val
		}
		appNotif.Subject = fmt.Sprintf("New comment added in %s `%s`", util.LowerSinglarize(appNotif.BaseEntityName), appNotif.BaseItemName)
	}

	duplicateMasker := make(map[string]bool, 0)
	//Send email/firebase notification to assignees/followers/creators
	for _, assignee := range appNotif.Assignees {
		if _, ok := duplicateMasker[assignee.UserID]; !ok {
			appNotif.Send(ctx, assignee, notificationType, db, firebaseSDKPath)
			duplicateMasker[assignee.UserID] = true
		}
	}

	for _, follower := range appNotif.Followers {
		if _, ok := duplicateMasker[follower.UserID]; !ok {
			appNotif.Send(ctx, follower, notificationType, db, firebaseSDKPath)
			duplicateMasker[follower.UserID] = true
		}
	}

	return appNotif.Save(ctx, notificationType, db)
}
