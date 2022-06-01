package forms

import (
	"encoding/json"

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

	accessFieldID := uuid.New().String()
	accessField := entity.Field{
		Key:         accessFieldID,
		Name:        "access_map",
		DisplayName: "Access",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
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
			Key:      "element",
			DataType: entity.TypeString,
		},
	}

	//V_E_C is view,edit,create
	accessMap := make(map[string]string, 0)
	accessMap["W"] = "V_E_C"
	accessMap["D"] = "V_E_C"
	fieldsBytes, _ := json.Marshal(accessMap)

	ownerVals := map[string]interface{}{
		userIDFieldID: currentUserID,
		nameFieldID:   name,
		avatarFieldID: avatar,
		emailFieldID:  email,
		teamFieldID:   []interface{}{teamID},
		accessFieldID: string(fieldsBytes),
	}

	return []entity.Field{userIDField, nameField, avatarField, emailField, teamListField, accessField}, ownerVals
}
