package notification

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/user"
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

func HostN(accName, input string) string {
	return Hostname(accName, input)
}

func WelcomeInvitation(draftID string, apps []string, accountName, host, requester, usrName, usrEmail string, db *sqlx.DB, sdb *database.SecDB) error {
	ctx := context.Background()
	app := "base"
	for _, ap := range apps {
		if ap != "base" {
			app = ap
		}
	}

	workBaseDomain := Hostname(accountName, host)
	magicLink, err := auth.CreateMagicLaunchLink(app, workBaseDomain, draftID, accountName, usrEmail, sdb)
	if err != nil {
		log.Println("***>***> WelcomeInvitation: unexpected/unhandled error occurred when creating the magic link. error:", err)
		return err
	}

	emailNotif := EmailNotification{
		AccountID:   draftID,
		To:          []interface{}{usrEmail},
		Subject:     "Welcome! Get started with workbaseONE",
		Name:        requester,
		Requester:   requester,
		AccountName: accountName,
		MagicLink:   magicLink,
	}
	return emailNotif.Send(ctx, TypeWelcome, db)
}

func JoinInvitation(accountID, accountName, accDomain string, teams []string, requester, usrName, usrEmail string, memberID string, db *sqlx.DB, sdb *database.SecDB) error {
	ctx := context.Background()

	magicLink, err := auth.CreateMagicLink(accountID, accDomain, usrName, usrEmail, memberID, sdb)
	if err != nil {
		log.Println("***>***> JoinInvitation: unexpected/unhandled error occurred when creating the magic link. error:", err)
		return err
	}

	emailNotif := EmailNotification{
		AccountID:   accountID,
		To:          []interface{}{usrEmail},
		Subject:     fmt.Sprintf("Invitation to join %s account", accountName),
		Name:        usrName,
		Requester:   requester,
		AccountName: accountName,
		MagicLink:   magicLink,
	}
	return emailNotif.Send(ctx, TypeMemberInvitation, db)
}

func VisitorInvitation(accountID, visitorID, body string, db *sqlx.DB, sdb *database.SecDB) error {
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
	magicLink, err := auth.CreateVisitorMagicLink(accountID, a.Domain, usrName, v.Email, visitorID, v.Token, sdb)
	if err != nil {
		log.Println("***>***> VisitorInvitation: unexpected/unhandled error occurred when creating the magic link. error:", err)
		return err
	}

	emailNotif := EmailNotification{
		AccountID:   accountID,
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

func OnAnItemLevelEvent(ctx context.Context, usrID string, entityCategory int, entityDisName, accountID, accDomain, teamID, entityID, itemID string, itemCreatorID *string, itemUpdatedAt int64, valueAddedFields []entity.Field, dirtyFields map[string]interface{}, source map[string][]string, notificationType NotificationType, db *sqlx.DB, sdb *database.SecDB, firebaseSDKPath string) (*item.Item, error) {
	appNotif := appNotificationBuilder(ctx, accountID, accDomain, teamID, usrID, entityCategory, entityID, itemID, itemCreatorID, valueAddedFields, dirtyFields, source, db, sdb)
	appNotif.CreatedAt = itemUpdatedAt

	if entityCategory == entity.CategoryEvent { //stop notifications for events
		return nil, nil
	}

	switch notificationType {
	case TypeReminder:
		if val, exist := appNotif.DirtyFields[entity.WhoDueBy]; exist {
			appNotif.Subject = fmt.Sprintf("%s is due on %s", util.UpperSinglarize(entityDisName), val)
		}
		appNotif.Body = fmt.Sprintf("%s", appNotif.Title)
	case TypeCreated:
		switch entityCategory {
		case entity.CategoryEmail:
			appNotif.Subject = fmt.Sprintf("An e-mail sent/received")
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
			appNotif.AddMembers(ctx, accountID, db)
			appNotif.Subject = fmt.Sprintf("New member added to your account")
			appNotif.Body = fmt.Sprintf("%s added to your account", util.UpperSinglarize(appNotif.Title))
		default:
			if len(source) == 0 { // add all members for the main module addition...
				appNotif.AddMembers(ctx, accountID, db)
			}

			appNotif.Subject = fmt.Sprintf("New %s created", util.LowerSinglarize(entityDisName))
			// enriching subject with base elements
			if appNotif.BaseEntityName != "" {
				appNotif.Subject = fmt.Sprintf("%s in %s", appNotif.Subject, util.LowerSinglarize(appNotif.BaseEntityName))
				if appNotif.BaseItemName != "" {
					appNotif.Subject = fmt.Sprintf("%s `%s`", appNotif.Subject, appNotif.BaseItemName)
				}
			}
			appNotif.Body = fmt.Sprintf("%s. %s", appNotif.Subject, appNotif.Title)
		}

	case TypeUpdated:
		appNotif.Subject = appNotif.Title
		switch entityCategory {
		case entity.CategoryUsers:
			appNotif.Subject = "Newly invited member logged in for the first time"
			appNotif.Body = fmt.Sprintf("`%s` logged in for the first time", appNotif.Title)
		default:
			// if val, exist := appNotif.DirtyFields[entity.WhoAssignee]; exist {
			// 	appNotif.Body = fmt.Sprintf("%s `%s` assigned to %s", util.UpperSinglarize(entityDisName), appNotif.Title, val)
			// }

			// if val, exist := appNotif.DirtyFields[entity.WhoDueBy]; exist {
			// 	body := fmt.Sprintf("%s `%s` due date set to %s", util.UpperSinglarize(entityDisName), appNotif.Title, val)
			// 	appNotif.Body = fmt.Sprintf("%s %s", appNotif.Body, body)
			// }

			if val, exist := appNotif.DirtyFields["modified_fields"]; exist {
				// appNotif.Body = fmt.Sprintf("%s `%s` updated with %s", util.UpperSinglarize(entityDisName), appNotif.Title, val)
				log.Println("modified vals ", val)
				log.Printf("valueAddedFields %+v", valueAddedFields)
				appNotif.Body = val
			}
		}

	case TypeChatConversationAdded:
		if val, exist := appNotif.DirtyFields[entity.WhoMessage]; exist {
			appNotif.Body = val
		}
		appNotif.Subject = fmt.Sprintf("New comment added in %s `%s`", util.LowerSinglarize(appNotif.BaseEntityName), appNotif.BaseItemName)
	}

	userSettingsMap, err := notificationSettings(ctx, accountID, appNotif.Assignees, appNotif.Followers, db)
	if err != nil {
		log.Println("***>***> OnAnItemLevelEvent: unexpected/unhandled error occurred when retriving userSettingsMap. error:", err)
	}
	duplicateMasker := make(map[string]bool, 0)
	//Send email/firebase notification to assignees/followers/creators
	for _, assignee := range appNotif.Assignees {
		log.Printf("assignee -- %+v", assignee)
		if _, ok := duplicateMasker[assignee.UserID]; !ok {
			if notifSettingMap, ok := userSettingsMap[assignee.UserID]; ok {
				if notifSettingMap[user.NSAssigned] == "true" {
					appNotif.Send(ctx, assignee, notificationType, db, firebaseSDKPath)
				}
			} else { // go with default flow if user settings not exist
				appNotif.Send(ctx, assignee, notificationType, db, firebaseSDKPath)
			}
			duplicateMasker[assignee.UserID] = true
		}
	}

	for _, follower := range appNotif.Followers {
		log.Printf("follower -- %+v", follower)
		if _, ok := duplicateMasker[follower.UserID]; !ok {
			if notifSettingMap, ok := userSettingsMap[follower.UserID]; ok {
				if (notificationType == TypeCreated && notifSettingMap[user.NSCreated] == "true") || (notificationType == TypeUpdated && notifSettingMap[user.NSUpdated] == "true") {
					appNotif.Send(ctx, follower, notificationType, db, firebaseSDKPath)
				}
			} else { // go with default flow if user settings not exist
				appNotif.Send(ctx, follower, notificationType, db, firebaseSDKPath)
			}
			duplicateMasker[follower.UserID] = true
		}
	}

	return appNotif.Save(ctx, notificationType, db)
}

func notificationSettings(ctx context.Context, accountID string, assignees, followers []entity.UserEntity, db *sqlx.DB) (map[string]map[string]string, error) {
	userMap := make(map[string]map[string]string, 0)
	userIDs := make([]interface{}, 0)
	for _, assignee := range assignees {
		userIDs = append(userIDs, assignee.UserID)
	}
	for _, follower := range followers {
		userIDs = append(userIDs, follower.UserID)
	}
	users, err := user.BulkRetrieveUserSetting(ctx, db, accountID, userIDs)
	if err != nil {
		return userMap, err
	}

	for _, u := range users {
		userMap[u.UserID] = user.UnmarshalNotificationSettings(u.NotificationSetting)
	}
	return userMap, nil
}

func Hostname(accName, input string) string {
	log.Println("hostname input ", input)
	url, err := url.Parse(input)
	if err != nil {
		log.Printf("***> unexpected error occurred when starting the trail. error: %v\n", err)
		return "app.workbaseone.com"
	}
	hostname := url.Hostname()
	for _, subDomain := range account.ExistingSubDomains {
		hostname = strings.TrimPrefix(url.Hostname(), fmt.Sprintf("%s.", subDomain))
	}
	log.Println(" hostname:----: ", hostname)
	//firebase is not supporting account as host.. using the workbaseone
	//return fmt.Sprintf("%s.%s", util.AccountAsHost(accName), hostname)
	return "app.workbaseone.com"
}
