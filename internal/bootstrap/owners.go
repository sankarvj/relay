package bootstrap

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func ownerFields(userID, name, avatar, email string) ([]entity.Field, map[string]interface{}) {
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
	}

	avatarFieldID := uuid.New().String()
	avatarField := entity.Field{
		Key:         avatarFieldID,
		Name:        "avatar",
		DisplayName: "Avatar",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	emailFieldID := uuid.New().String()
	emailField := entity.Field{
		Key:         emailFieldID,
		Name:        "email",
		DisplayName: "Email",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	gTokenFieldID := uuid.New().String()
	tokenField := entity.Field{
		Key:         gTokenFieldID,
		Name:        "gtoken",
		DisplayName: "",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
	}

	ownerVals := map[string]interface{}{
		userIDFieldID: userID,
		nameFieldID:   name,
		avatarFieldID: avatar,
		emailFieldID:  email,
		gTokenFieldID: "",
	}

	return []entity.Field{userIDField, nameField, avatarField, emailField, tokenField}, ownerVals
}
