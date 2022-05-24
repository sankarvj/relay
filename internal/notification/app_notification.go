package notification

import (
	"context"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

type AppNotification struct {
	AccountID string
	TeamID    string
	UserID    string
	EntityID  string
	ItemID    string
	Subject   string
	Body      string
	UserName  string
	Followers []entity.UserEntity
	Assignees []entity.UserEntity
	Specifics
}

type Specifics struct {
	Title          string
	DueBy          string
	Names          string
	ModifiedFields []string
}

func appNotificationBuilder(ctx context.Context, accountID, teamID, userID, entityID, itemID string, itemCreatorID *string, valueAddedFields []entity.Field, dirtyFields map[string]interface{}, db *sqlx.DB) AppNotification {

	appNotif := AppNotification{
		AccountID: accountID,
		TeamID:    teamID,
		UserID:    userID,
		EntityID:  entityID,
		ItemID:    itemID,
	}

	usr, err := user.RetrieveUser(ctx, db, userID)
	if err == nil {
		appNotif.UserName = *usr.Name
	}

	appNotif.Assignees = make([]entity.UserEntity, 0)
	appNotif.Followers = make([]entity.UserEntity, 0)
	//subject, body, userItem should be populated here
	appNotif.ModifiedFields = []string{}
	for _, f := range valueAddedFields {

		if f.Value == nil {
			continue
		}

		if f.IsTitleLayout() {
			appNotif.Title = f.Value.(string)
		}

		if f.Who == entity.WhoDueBy && f.DataType == entity.TypeDateTime {
			if _, ok := dirtyFields[f.Key]; ok {
				when, _ := util.ParseTime(f.Value.(string))
				appNotif.DueBy = util.FormatTimeView(when)
			}
		}

		if f.Who == entity.WhoAssignee {
			assignees := dirtyFields[f.Key].([]interface{})
			names := []string{}
			for _, assignee := range assignees {
				userItem, err := entity.RetriveUserItem(ctx, accountID, assignee.(string), db)
				if err != nil {
					log.Println("***>***> ItemUpdates: unexpected/unhandled error occurred while retriving userItem from memberID. error:", err)
					continue
				}
				appNotif.Assignees = append(appNotif.Assignees, *userItem)
				names = append(names, userItem.Name)
			}
			appNotif.Names = strings.Join(names, ",")
		}

		if _, ok := dirtyFields[f.Key]; ok {
			appNotif.ModifiedFields = append(appNotif.ModifiedFields, f.DisplayName)
		}
	}

	return appNotif
}

func (appNotif AppNotification) Send(ctx context.Context, notifType NotificationType, db *sqlx.DB) error {
	notificationItem := entity.NotificationEntityItem{
		AccountID: appNotif.AccountID,
		TeamID:    appNotif.TeamID,
		EntityID:  appNotif.EntityID,
		ItemID:    appNotif.ItemID,
		UserID:    appNotif.UserID,
		UserName:  appNotif.UserName,
		Subject:   appNotif.Subject,
		Body:      appNotif.Body,
		Followers: fetchMemberIDs(appNotif.Followers),
		Assignees: fetchMemberIDs(appNotif.Assignees),
		Type:      int(notifType),
	}

	_, err := entity.SaveFixedEntityItem(ctx, appNotif.AccountID, appNotif.TeamID, appNotif.UserID, entity.FixedEntityNotification, "Notification", "", "", util.ConvertInterfaceToMap(notificationItem), db)
	if err != nil {
		return err
	}

	return nil
}

func fetchMemberIDs(entities []entity.UserEntity) []string {
	ids := make([]string, 0)
	for _, e := range entities {
		ids = append(ids, e.UserID)
	}
	return ids
}

func fetchUserIDs(users []user.User) []string {
	ids := make([]string, 0)
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	return ids
}
