package notification

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/rule/engine"
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
	BaseIds   []string
	Due       time.Time
	Specifics
}

type Specifics struct {
	Title       string
	DirtyFields map[string]string
}

func appNotificationBuilder(ctx context.Context, accountID, teamID, userID, entityID, itemID string, itemCreatorID *string, valueAddedFields []entity.Field, dirtyFields map[string]interface{}, baseIds []string, db *sqlx.DB) AppNotification {

	appNotif := AppNotification{
		AccountID: accountID,
		TeamID:    teamID,
		UserID:    userID,
		EntityID:  entityID,
		ItemID:    itemID,
		Followers: make([]entity.UserEntity, 0),
		Assignees: make([]entity.UserEntity, 0),
		BaseIds:   baseIds,
	}
	appNotif.DirtyFields = make(map[string]string, 0)

	if itemCreatorID != nil && *itemCreatorID != engine.UUID_SYSTEM_USER {
		creator, err := user.RetrieveUser(ctx, db, *itemCreatorID)
		if err != nil {
			log.Println("***>***> appNotificationBuilder: unexpected/unhandled error occurred while retriving user from creatorID. error:", err)
		} else {
			if memberID, ok := creator.AccountsB()[accountID]; ok {
				userItem, err := entity.RetriveUserItem(ctx, accountID, memberID.(string), db)
				if err != nil {
					log.Println("***>***> appNotificationBuilder: unexpected/unhandled error occurred while retriving userItem from memberID. error:", err)
				} else {
					appNotif.Followers = append(appNotif.Followers, *userItem)
				}
			}
		}
	}

	usr, err := user.RetrieveUser(ctx, db, userID)
	if err == nil {
		appNotif.UserName = *usr.Name
	}

	//subject, body, userItem should be populated here
	modifiedFields := make([]string, 0)
	for _, f := range valueAddedFields {

		if f.Value == nil {
			continue
		}

		if _, ok := dirtyFields[f.Key]; ok {
			modifiedFields = append(modifiedFields, f.DisplayName)
		}

		if f.IsTitleLayout() {
			appNotif.Title = f.Value.(string)
		}

		if f.Who == entity.WhoDueBy && f.DataType == entity.TypeDateTime {
			when, _ := util.ParseTime(f.Value.(string))
			appNotif.Due = when
			if _, ok := dirtyFields[f.Key]; ok {
				appNotif.DirtyFields[entity.WhoDueBy] = util.FormatTimeView(when)
			}
		}

		if f.Who == entity.WhoAssignee {
			for _, assignee := range f.Value.([]interface{}) {
				userItem, err := entity.RetriveUserItem(ctx, accountID, assignee.(string), db)
				if err != nil {
					log.Println("***>***> ItemUpdates: unexpected/unhandled error occurred while retriving userItem from memberID. error:", err)
					continue
				}
				appNotif.Assignees = append(appNotif.Assignees, *userItem)
			}

			if _, ok := dirtyFields[f.Key]; ok {
				appNotif.DirtyFields[entity.WhoAssignee] = appNotif.assigneeNames()
			}
		}

	}
	appNotif.DirtyFields["modified_fields"] = strings.Join(modifiedFields, ",")

	return appNotif
}

func (appNotif AppNotification) Send(ctx context.Context, notifType NotificationType, db *sqlx.DB) (*item.Item, error) {
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
		BaseIds:   appNotif.BaseIds,
		Type:      int(notifType),
	}

	it, err := entity.SaveFixedEntityItem(ctx, appNotif.AccountID, appNotif.TeamID, appNotif.UserID, entity.FixedEntityNotification, "Notification", "", "", util.ConvertInterfaceToMap(notificationItem), db)
	if err != nil {
		return nil, err
	}

	return &it, nil
}

func fetchMemberIDs(entities []entity.UserEntity) []string {
	ids := make([]string, 0)
	for _, e := range entities {
		ids = append(ids, e.MemberID)
	}
	return ids
}

func fetchUserIDs(entities []entity.UserEntity) []string {
	ids := make([]string, 0)
	for _, e := range entities {
		ids = append(ids, e.UserID)
	}
	return ids
}

func (appNotif AppNotification) assigneeNames() string {
	names := make([]string, 0)
	for _, ass := range appNotif.Assignees {
		names = append(names, ass.Name)
	}
	return strings.Join(names[:], ",")
}
