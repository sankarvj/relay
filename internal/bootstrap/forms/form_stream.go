package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func StreamFields(ownerEntityID, ownerEntitySearchKey string) []entity.Field {
	titleFieldID := uuid.New().String()
	titleField := entity.Field{
		Key:         titleFieldID,
		Name:        "title",
		DisplayName: "Title",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	messageFieldID := uuid.New().String()
	messageField := entity.Field{
		Key:         messageFieldID,
		Name:        "label",
		DisplayName: "Label",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle},
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
		Meta:        map[string]string{entity.MetaKeyDisplayGex: ownerEntitySearchKey, entity.MetaKeyLayout: entity.MetaLayoutUsers},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{titleField, messageField, followerField}
}

func StreamVals(streamEntity entity.Entity, title, message, file string) map[string]interface{} {
	streamVals := map[string]interface{}{
		"title": title,
		"label": message,
	}
	return keyMap(streamEntity.NamedKeys(), streamVals)
}
