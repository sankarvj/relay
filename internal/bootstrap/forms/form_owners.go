package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func OwnerFields(teamID, currentUserID, name, avatar, email string) ([]entity.Field, map[string]interface{}) {

	userIDFieldID := uuid.New().String()
	userIDField := entity.Field{
		Key:         userIDFieldID,
		Name:        "user_id",
		DisplayName: "User ID",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
		Who:         entity.WhoIdentifier,
	}

	nameFieldID := uuid.New().String()
	nameField := entity.Field{
		Key:         nameFieldID,
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	avatarFieldID := uuid.New().String()
	avatarField := entity.Field{
		Key:         avatarFieldID,
		Name:        "avatar",
		DisplayName: "Avatar",
		DomType:     entity.DomImage,
		DataType:    entity.TypeString,
		Who:         entity.WhoAvatar,
	}

	emailFieldID := uuid.New().String()
	emailField := entity.Field{
		Key:         emailFieldID,
		Name:        "email",
		DisplayName: "Email",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	roleFieldID := uuid.New().String()
	roleField := entity.Field{
		Key:         roleFieldID,
		Name:        "role",
		DisplayName: "Role",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeList,
		Choices: []entity.Choice{
			{
				ID:           "ADMIN",
				DisplayValue: "ADMIN",
			},
			{
				ID:           "USER",
				DisplayValue: "USER",
			},
			{
				ID:           "MEMBER",
				DisplayValue: "MEMBER",
			},
			{
				ID:           "VISITOR",
				DisplayValue: "VISITOR",
			},
		},
		Field: &entity.Field{
			Key:      "id",
			DataType: entity.TypeString,
		},
	}

	teamFieldID := uuid.New().String()
	teamListField := entity.Field{
		Key:         teamFieldID,
		Name:        "team_ids",
		DisplayName: "Associated teams",
		DataType:    entity.TypeList,
		DomType:     entity.DomMultiSelect,
		Meta:        map[string]string{entity.MetaMultiChoice: "true"},
		Field: &entity.Field{
			Key:      "id",
			DataType: entity.TypeString,
		},
	}

	ownerVals := map[string]interface{}{
		userIDFieldID: currentUserID,
		nameFieldID:   name,
		avatarFieldID: avatar,
		emailFieldID:  email,
		teamFieldID:   []interface{}{teamID},
		roleFieldID:   []interface{}{"ADMIN"},
	}

	return []entity.Field{userIDField, nameField, avatarField, emailField, roleField, teamListField}, ownerVals
}
