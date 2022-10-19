package notification

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

type AppNotification struct {
	AccountID      string
	AccountDomain  string
	TeamID         string
	UserID         string
	EntityID       string
	ItemID         string
	Subject        string
	Body           string
	UserName       string
	UserAvatar     string
	CreatedAt      int64
	Followers      []entity.UserEntity
	Assignees      []entity.UserEntity
	BaseIds        []string //useful for fetching the events....
	BaseEntityName string
	BaseItemName   string
	Due            time.Time
	Specifics
}

type Specifics struct {
	Title       string
	DirtyFields map[string]string
}

func appNotificationBuilder(ctx context.Context, accountID, accountDomain, teamID, userID, entityID, itemID string, itemCreatorID *string, valueAddedFields []entity.Field, dirtyFields map[string]interface{}, source map[string][]string, db *sqlx.DB) AppNotification {
	appNotif := AppNotification{
		AccountID:     accountID,
		AccountDomain: accountDomain,
		TeamID:        teamID,
		UserID:        userID,
		EntityID:      entityID,
		ItemID:        itemID,
		Followers:     make([]entity.UserEntity, 0),
		Assignees:     make([]entity.UserEntity, 0),
		BaseIds:       make([]string, 0), //events filter use case. check README for more info
	}

	for baseEntityID, baseItemIDs := range source {
		baseEntity, berr := entity.Retrieve(ctx, accountID, baseEntityID, db)
		appNotif.BaseEntityName = baseEntity.DisplayName
		for i, baseItemID := range baseItemIDs {
			//making baseID for filtering notifications per item (events).
			appNotif.BaseIds = append(appNotif.BaseIds, fmt.Sprintf("%s#%s", baseEntityID, baseItemID))
			if i == 0 && berr == nil { // for now fetching one time is enough
				it, err := item.Retrieve(ctx, baseEntityID, baseItemID, db)
				if err == nil && it.State != item.StateWebForm {
					titleField := entity.TitleField(baseEntity.FieldsIgnoreError())
					appNotif.BaseItemName = it.Fields()[titleField.Key].(string)

					//adding base item follower and assignees
					appNotif.AddFollower(ctx, accountID, it.UserID, db)
					baseValueAddedFields := baseEntity.ValueAdd(it.Fields())
					for _, f := range baseValueAddedFields {
						if f.Value == nil {
							continue
						}
						if f.Who == entity.WhoAssignee {
							appNotif.AddAssignees(ctx, accountID, f.RefID, f.Value.([]interface{}), db)
						}
					}
				}
			}
		}
	}

	appNotif.DirtyFields = make(map[string]string, 0)
	if itemCreatorID != nil && *itemCreatorID != user.UUID_SYSTEM_USER && *itemCreatorID != userID {
		appNotif.AddFollower(ctx, accountID, itemCreatorID, db)
	}

	switch userID {
	case user.UUID_ENGINE_USER:
		appNotif.UserName = "automation workflow"
		appNotif.UserAvatar = "https://avatars.dicebear.com/api/bottts/workflow.svg"
	case user.UUID_SYSTEM_USER:
		appNotif.UserName = "system"
		appNotif.UserAvatar = "https://avatars.dicebear.com/api/bottts/system.svg"
	default:
		usr, err := user.RetrieveUser(ctx, db, userID)
		if err == nil {
			appNotif.UserName = *usr.Name
			appNotif.UserAvatar = *usr.Avatar
		}
	}

	//subject, body, userItem should be populated here
	modifiedFields := make([]string, 0)
	for _, f := range valueAddedFields {

		if f.Value == nil {
			continue
		}

		if _, ok := dirtyFields[f.Key]; ok {
			modifiedFields = append(modifiedFields, fmt.Sprintf("%s", f.DisplayName))
		}

		if f.IsTitleLayout() {
			appNotif.Title = util.TruncateText(f.Value.(string), 30)
		}

		if f.Who == entity.WhoDueBy && f.DataType == entity.TypeDateTime {
			when, _ := util.ParseTime(f.Value.(string))
			appNotif.Due = when
			if _, ok := dirtyFields[f.Key]; ok {
				appNotif.DirtyFields[entity.WhoDueBy] = util.FormatTimeView(when)
			}
		}

		if f.Who == entity.WhoAssignee {
			appNotif.AddAssignees(ctx, accountID, f.RefID, f.Value.([]interface{}), db)
			if _, ok := dirtyFields[f.Key]; ok {
				appNotif.DirtyFields[entity.WhoAssignee] = appNotif.assigneeNames()
			}
		}
		if f.Who == entity.WhoFollower {
			appNotif.AddFollowers(ctx, accountID, f.RefID, f.Value.([]interface{}), db)
			if _, ok := dirtyFields[f.Key]; ok {
				appNotif.DirtyFields[entity.WhoAssignee] = appNotif.assigneeNames()
			}
		}
	}
	appNotif.DirtyFields["modified_fields"] = strings.Join(modifiedFields, ",")

	//adding message for chat conversation if exists
	if val, exist := dirtyFields[entity.WhoMessage]; exist {
		appNotif.DirtyFields[entity.WhoMessage] = val.(string)
	}

	return appNotif
}

func (appNotif *AppNotification) AddAssignees(ctx context.Context, accountID, assigneeEntityID string, assignees []interface{}, db *sqlx.DB) {
	ownerEntity, err := entity.Retrieve(ctx, accountID, assigneeEntityID, db)
	if err != nil {
		log.Println("***>***> ItemUpdates: unexpected/unhandled error occurred while retriving owner entity. error:", err)
	}
	for _, assignee := range assignees {
		userItem, err := entity.RetriveUserItem(ctx, accountID, ownerEntity.ID, assignee.(string), db)
		if err != nil {
			log.Println("***>***> ItemUpdates: unexpected/unhandled error occurred while retriving userItem from memberID. error:", err)
			continue
		}
		appNotif.Assignees = append(appNotif.Assignees, *userItem)
	}
}

func (appNotif *AppNotification) AddFollowers(ctx context.Context, accountID, assigneeEntityID string, assignees []interface{}, db *sqlx.DB) {
	ownerEntity, err := entity.Retrieve(ctx, accountID, assigneeEntityID, db)
	if err != nil {
		log.Println("***>***> ItemUpdates: unexpected/unhandled error occurred while retriving owner entity. error:", err)
	}
	for _, assignee := range assignees {
		userItem, err := entity.RetriveUserItem(ctx, accountID, ownerEntity.ID, assignee.(string), db)
		if err != nil {
			log.Println("***>***> ItemUpdates: unexpected/unhandled error occurred while retriving userItem from memberID. error:", err)
			continue
		}
		appNotif.Followers = append(appNotif.Followers, *userItem)
	}
}

func (appNotif *AppNotification) AddFollower(ctx context.Context, accountID string, itemCreatorID *string, db *sqlx.DB) {
	creator, err := user.RetrieveUser(ctx, db, *itemCreatorID)
	if err != nil {
		log.Println("***>***> appNotificationBuilder: unexpected/unhandled error occurred while retriving user from creatorID. error:", err)
	} else {
		ownerEntity, err := entity.RetrieveFixedEntity(ctx, db, accountID, "", entity.FixedEntityOwner)
		if err != nil {
			log.Println("***>***> appNotificationBuilder: unexpected/unhandled error occurred while retriving owner entity. error:", err)
		}
		if memberID, ok := creator.AccountsB()[accountID]; ok {
			userItem, err := entity.RetriveUserItem(ctx, accountID, ownerEntity.ID, memberID.(string), db)
			if err != nil {
				log.Println("***>***> appNotificationBuilder: unexpected/unhandled error occurred while retriving userItem from memberID. error:", err)
			} else {
				appNotif.Followers = append(appNotif.Followers, *userItem)
			}
		}
	}
}

func (appNotif *AppNotification) AddMoreFollowers(ctx context.Context, accountID string, db *sqlx.DB) {
	ownerEntity, err := entity.RetrieveFixedEntity(ctx, db, accountID, "", entity.FixedEntityOwner)
	if err != nil {
		log.Println("***>***> appNotificationBuilder: unexpected/unhandled error occurred while retriving owner entity. error:", err)
	}
	userItems, err := item.ListFilterByState(ctx, accountID, ownerEntity.ID, item.StateDefault, db)
	if err != nil {
		log.Println("***>***> appNotificationBuilder: unexpected/unhandled error occurred while retriving owner items. error:", err)
	}

	for _, userItem := range userItems {
		var userEntityItem entity.UserEntity
		valueAddedFields := ownerEntity.ValueAdd(userItem.Fields())
		err = entity.ParseFixedEntity(valueAddedFields, &userEntityItem)
		if err != nil {
			log.Println("***>***> appNotificationBuilder: unexpected/unhandled error occurred while parsing owner items. error:", err)
			continue
		}
		userEntityItem.MemberID = userItem.ID
		if userEntityItem.UserID != "" {
			appNotif.Followers = append(appNotif.Followers, userEntityItem)
		}
	}
}

func (appNotif AppNotification) Send(ctx context.Context, assignee entity.UserEntity, notificationType NotificationType, db *sqlx.DB, firebaseSDKPath string) (error, error) {
	emailNotif := EmailNotification{
		Name:      strings.Title(assignee.Name),
		To:        []interface{}{assignee.Email},
		Subject:   appNotif.Subject,
		Body:      appNotif.Body,
		MagicLink: auth.SimpleLink(appNotif.AccountID, appNotif.AccountDomain, appNotif.TeamID, appNotif.EntityID, appNotif.ItemID),
	}
	err1 := emailNotif.Send(ctx, notificationType, db)
	if err1 != nil {
		log.Println("***>***> emailNotif.Send. error:", err1)
	}

	fbNotif := FirebaseNotification{
		AccountID:    appNotif.AccountID,
		TargetUserID: assignee.UserID,
		UserName:     appNotif.UserName,
		UserAvatar:   appNotif.UserAvatar,
		CreatedAt:    util.ConvertMilliToTime(appNotif.CreatedAt),
		Subject:      appNotif.Subject,
		Body:         appNotif.Body,
		SDKPath:      firebaseSDKPath,
	}
	err2 := fbNotif.Send(ctx, notificationType, db)
	if err2 != nil {
		log.Println("***>***> fbNotif.Send. error:", err2)
	}
	return err1, err2
}

func (appNotif AppNotification) Save(ctx context.Context, notifType NotificationType, db *sqlx.DB) (*item.Item, error) {
	notificationItem := entity.NotificationEntityItem{
		AccountID:  appNotif.AccountID,
		TeamID:     appNotif.TeamID,
		EntityID:   appNotif.EntityID,
		ItemID:     appNotif.ItemID,
		UserID:     appNotif.UserID,
		UserName:   appNotif.UserName,
		UserAvatar: appNotif.UserAvatar,
		Subject:    appNotif.Subject,
		Body:       appNotif.Body,
		Followers:  fetchMemberIDs(appNotif.Followers),
		Assignees:  fetchMemberIDs(appNotif.Assignees),
		BaseIds:    appNotif.BaseIds,
		Type:       int(notifType),
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

// func fetchUserIDs(entities []entity.UserEntity) []string {
// 	ids := make([]string, 0)
// 	for _, e := range entities {
// 		ids = append(ids, e.UserID)
// 	}
// 	return ids
// }

func (appNotif AppNotification) assigneeNames() string {
	names := make([]string, 0)
	for _, ass := range appNotif.Assignees {
		names = append(names, ass.Name)
	}
	return strings.Join(names[:], ",")
}
