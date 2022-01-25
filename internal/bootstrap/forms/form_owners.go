package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func OwnerFields(userID, name, avatar, email string) ([]entity.Field, map[string]interface{}) {
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

	ownerVals := map[string]interface{}{
		userIDFieldID: userID,
		nameFieldID:   name,
		avatarFieldID: avatar,
		emailFieldID:  email,
	}

	return []entity.Field{userIDField, nameField, avatarField, emailField}, ownerVals
}
