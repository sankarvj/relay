package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func TeamFields(name, desc string) ([]entity.Field, map[string]interface{}) {

	nameFieldID := uuid.New().String()
	nameField := entity.Field{
		Key:         nameFieldID,
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	descriptionFieldID := uuid.New().String()
	descriptionField := entity.Field{
		Key:         descriptionFieldID,
		Name:        "description",
		DisplayName: "Description",
		DomType:     entity.DomImage,
		DataType:    entity.TypeString,
		Who:         entity.WhoAvatar,
	}

	teamVals := map[string]interface{}{
		nameFieldID:        name,
		descriptionFieldID: desc,
	}

	return []entity.Field{nameField, descriptionField}, teamVals
}
