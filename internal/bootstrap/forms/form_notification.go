package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func NotificationFields(ownerEntityID string) []entity.Field {
	subjectFieldID := uuid.New().String()
	subjectField := entity.Field{
		Key:         subjectFieldID,
		Name:        "subject",
		DisplayName: "Subject",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	bodyFieldID := uuid.New().String()
	bodyField := entity.Field{
		Key:         bodyFieldID,
		Name:        "body",
		DisplayName: "Body",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	typeFieldID := uuid.New().String()
	typeField := entity.Field{
		Key:         typeFieldID,
		Name:        "type",
		DisplayName: "Type",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	timeFieldID := uuid.New().String()
	timeField := entity.Field{
		Key:         timeFieldID,
		Name:        "created_at",
		DisplayName: "Created At",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	accountFieldID := uuid.New().String()
	accountField := entity.Field{
		Key:         accountFieldID,
		Name:        "account_id",
		DisplayName: "Account",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	teamID := uuid.New().String()
	teamField := entity.Field{
		Key:         teamID,
		Name:        "team_id",
		DisplayName: "Team",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	userFieldID := uuid.New().String()
	userField := entity.Field{
		Key:         userFieldID,
		Name:        "user_id",
		DisplayName: "User",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	userNameFieldID := uuid.New().String()
	userNameField := entity.Field{
		Key:         userNameFieldID,
		Name:        "user_name",
		DisplayName: "User Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	userAvatarFieldID := uuid.New().String()
	userAvatarField := entity.Field{
		Key:         userAvatarFieldID,
		Name:        "user_avatar",
		DisplayName: "User Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	entityFieldID := uuid.New().String()
	entityField := entity.Field{
		Key:         entityFieldID,
		Name:        "entity_id",
		DisplayName: "Entity",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	itemFieldID := uuid.New().String()
	itemField := entity.Field{
		Key:         itemFieldID,
		Name:        "item_id",
		DisplayName: "Item",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	followersFieldID := uuid.New().String()
	followerField := entity.Field{
		Key:         followersFieldID,
		Name:        "followers",
		DisplayName: "Followers",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       ownerEntityID,
		Who:         entity.WhoFollower,
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	ownersFieldID := uuid.New().String()
	ownerField := entity.Field{
		Key:         ownersFieldID,
		Name:        "assignees",
		DisplayName: "Assignees",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       ownerEntityID,
		Who:         entity.WhoAssignee,
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	baseFieldID := uuid.New().String()
	baseField := entity.Field{
		Key:         baseFieldID,
		Name:        "base_ids",
		DisplayName: "Base Ids",
		DataType:    entity.TypeList,
		DomType:     entity.DomMultiSelect,
		Field: &entity.Field{
			Key:      "id",
			DataType: entity.TypeString,
		},
	}

	return []entity.Field{subjectField, bodyField, typeField, timeField, accountField, teamField, entityField, userField, userNameField, userAvatarField, itemField, followerField, ownerField, baseField}
}
