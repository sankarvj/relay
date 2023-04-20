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
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
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
	Category       int
	Specifics
}

type Specifics struct {
	Title       string
	DirtyFields map[string]string
}

func appNotificationBuilder(ctx context.Context, accountID, accountDomain, teamID, userID string, entityCategory int, entityID, itemID string, itemCreatorID *string, valueAddedFields []entity.Field, dirtyFields map[string]interface{}, source map[string][]string, db *sqlx.DB, sdb *database.SecDB) AppNotification {
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
		Category:      entityCategory,
	}

	for baseEntityID, baseItemIDs := range source {
		baseEntity, berr := entity.Retrieve(ctx, accountID, baseEntityID, db, sdb)
		appNotif.BaseEntityName = baseEntity.DisplayName
		for i, baseItemID := range baseItemIDs {
			if baseEntityID != "" && baseItemID != "" {
				//making baseID for filtering notifications per item (events).
				appNotif.BaseIds = append(appNotif.BaseIds, fmt.Sprintf("%s#%s", baseEntityID, baseItemID))
				if i == 0 && berr == nil { // for now fetching one time is enough
					it, err := item.Retrieve(ctx, accountID, baseEntityID, baseItemID, db)
					if err == nil && it.State != item.StateWebForm {
						titleField := entity.TitleField(baseEntity.EasyFields())
						if it.Fields()[titleField.Key] != nil {
							appNotif.BaseItemName = it.Fields()[titleField.Key].(string)
						}

						//adding base item creator
						appNotif.AddCreators(ctx, accountID, it.UserID, db, sdb)
						//adding base item assignees
						baseValueAddedFields := baseEntity.ValueAdd(it.Fields())
						for _, f := range baseValueAddedFields {
							if f.Value == nil {
								continue
							}
							if f.Who == entity.WhoAssignee {
								appNotif.AddAssignees(ctx, accountID, f.RefID, f.Value.([]interface{}), db, sdb)
							}
							if f.Who == entity.WhoFollower {
								appNotif.AddFollowers(ctx, accountID, f.RefID, f.Value.([]interface{}), db, sdb)
							}
						}
					}
				}
			}
		}
	}

	if len(appNotif.BaseIds) == 0 {
		appNotif.BaseIds = append(appNotif.BaseIds, fmt.Sprintf("%s#%s", entityID, itemID))
	}

	appNotif.DirtyFields = make(map[string]string, 0)
	if itemCreatorID != nil && *itemCreatorID != user.UUID_SYSTEM_USER && *itemCreatorID != userID && *itemCreatorID != user.UUID_ENGINE_USER {
		appNotif.AddCreators(ctx, accountID, itemCreatorID, db, sdb)
	}

	switch userID {
	case user.UUID_ENGINE_USER:
		appNotif.UserName = "automation"
		appNotif.UserAvatar = "https://avatars.dicebear.com/api/bottts/workflow.svg"
	case user.UUID_SYSTEM_USER:
		appNotif.UserName = "system"
		appNotif.UserAvatar = "https://avatars.dicebear.com/api/bottts/system.svg"
	default:
		usr, err := user.RetrieveUser(ctx, db, accountID, userID)
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
			if f.IsReference() || f.IsList() {
				if f.Value != nil {
					vals := f.Value.([]interface{})
					var disp string
					for _, v := range vals {
						if v != nil {
							for _, choice := range f.Choices {
								if choice.ID == v.(string) {
									disp = fmt.Sprintf("%s %s", disp, choice.DisplayValue.(string))
								}
							}
						}
					}
					modifiedFields = append(modifiedFields, fmt.Sprintf("%s set to %s", f.DisplayName, disp))
				}

			} else {
				modifiedFields = append(modifiedFields, fmt.Sprintf("%s set to %s", f.DisplayName, f.Value))
			}
		}

		if f.IsTitleLayout() {
			appNotif.Title = util.TruncateText(f.Value.(string), 30)
		}

		if f.Who == entity.WhoAssignee {
			appNotif.AddAssignees(ctx, accountID, f.RefID, f.Value.([]interface{}), db, sdb)
		}
		if f.Who == entity.WhoFollower {
			appNotif.AddFollowers(ctx, accountID, f.RefID, f.Value.([]interface{}), db, sdb)
		}
	}
	appNotif.DirtyFields["modified_fields"] = strings.Join(modifiedFields, ",")

	//adding message for chat conversation if exists
	if val, exist := dirtyFields[entity.WhoMessage]; exist {
		appNotif.DirtyFields[entity.WhoMessage] = val.(string)
	}

	return appNotif
}

func (appNotif *AppNotification) AddAssignees(ctx context.Context, accountID, ownerEntityID string, ownerIds []interface{}, db *sqlx.DB, sdb *database.SecDB) {
	ownerEntity, err := entity.Retrieve(ctx, accountID, ownerEntityID, db, sdb)
	if err != nil {
		log.Println("***>***> appNotificationBuilder: unexpected/unhandled error occurred while retriving owner entity in add_assigness. error:", err)
	}
	owners, err := users(ctx, accountID, ownerEntity, ownerIds, db)
	if err != nil {
		log.Println("***>***> appNotificationBuilder: unexpected/unhandled error occurred while retriving owners in add_assigness. error:", err)
	}
	appNotif.Assignees = append(appNotif.Assignees, owners...)
}

func (appNotif *AppNotification) AddFollowers(ctx context.Context, accountID, ownerEntityID string, ownerIds []interface{}, db *sqlx.DB, sdb *database.SecDB) {
	ownerEntity, err := entity.Retrieve(ctx, accountID, ownerEntityID, db, sdb)
	if err != nil {
		log.Println("***>***> appNotificationBuilder: unexpected/unhandled error occurred while retriving owner entity in add_followers. error:", err)
	}
	owners, err := users(ctx, accountID, ownerEntity, ownerIds, db)
	if err != nil {
		log.Println("***>***> appNotificationBuilder: unexpected/unhandled error occurred while retriving owners in add_followers. error:", err)
	}
	appNotif.Followers = append(appNotif.Followers, owners...)
}

func (appNotif *AppNotification) AddCreators(ctx context.Context, accountID string, itemCreatorID *string, db *sqlx.DB, sdb *database.SecDB) {
	creator, err := user.RetrieveUser(ctx, db, accountID, *itemCreatorID)
	if err != nil {
		log.Println("***>***> appNotificationBuilder: unexpected/unhandled error occurred while retriving user from creatorID. error:", err)
	} else {
		ownerEntity, err := entity.RetrieveFixedEntity(ctx, db, accountID, "", entity.FixedEntityOwner)
		if err != nil {
			log.Println("***>***> appNotificationBuilder: unexpected/unhandled error occurred while retriving owner entity. error:", err)
		}
		userItem, err := entity.RetriveUserItem(ctx, accountID, ownerEntity.ID, creator.MemberID, db, sdb)
		if err != nil {
			log.Println("***>***> appNotificationBuilder: unexpected/unhandled error occurred while retriving userItem from memberID. error:", err)
		} else {
			appNotif.Followers = append(appNotif.Followers, *userItem)
		}
	}
}

func (appNotif *AppNotification) AddMembers(ctx context.Context, accountID string, db *sqlx.DB) {
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

		if userEntityItem.UserID != "" {
			userEntityItem.MemberID = userItem.ID
			appNotif.Followers = append(appNotif.Followers, userEntityItem)
		}
	}
}

func (appNotif AppNotification) Send(ctx context.Context, assignee entity.UserEntity, notificationType NotificationType, db *sqlx.DB, firebaseSDKPath string) (error, error) {
	emailNotif := EmailNotification{
		AccountID: appNotif.AccountID,
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
		Category:   appNotif.Category,
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

func (appNotif AppNotification) followerNames() string {
	names := make([]string, 0)
	for _, ass := range appNotif.Followers {
		names = append(names, ass.Name)
	}
	return strings.Join(names[:], ",")
}

func users(ctx context.Context, accountID string, ownerEntity entity.Entity, owners []interface{}, db *sqlx.DB) ([]entity.UserEntity, error) {
	users := make([]entity.UserEntity, 0)
	items, err := item.BulkRetrieveItems(ctx, accountID, owners, db)
	if err != nil {
		return users, err
	}

	for _, it := range items {
		ownerValueAdded := ownerEntity.ValueAdd(it.Fields())
		var userEntityItem entity.UserEntity
		err = entity.ParseFixedEntity(ownerValueAdded, &userEntityItem)
		if err != nil {
			return users, err
		}
		userEntityItem.MemberID = it.ID
		users = append(users, userEntityItem)
	}
	return users, nil

}
