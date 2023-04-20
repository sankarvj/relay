package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func NotifyFields(ownerEntityID, ownerEntitySearchKey string) []entity.Field {
	titleFieldID := uuid.New().String()
	titleField := entity.Field{
		Key:         titleFieldID,
		Name:        "title",
		DisplayName: "Title",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	ownersFieldID := uuid.New().String()
	ownerField := entity.Field{
		Key:         ownersFieldID,
		Name:        "owner",
		DisplayName: "Owner",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       ownerEntityID,
		Who:         entity.WhoAssignee,
		Meta:        map[string]string{entity.MetaMultiChoice: "false", entity.MetaKeyDisplayGex: ownerEntitySearchKey, entity.MetaKeyLayout: entity.MetaLayoutUsers},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{titleField, ownerField}
}

func NotifyVals(notifyEntity entity.Entity, title string) map[string]interface{} {
	streamVals := map[string]interface{}{
		"title": title,
	}
	return keyMap(notifyEntity.NameKeyMapWrapper(), streamVals)
}
