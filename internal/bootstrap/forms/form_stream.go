package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func StreamFields() []entity.Field {
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
		Name:        "message",
		DisplayName: "Message",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle},
	}

	return []entity.Field{titleField, messageField}
}

func StreamVals(streamEntity entity.Entity, title, message, file string) map[string]interface{} {
	streamVals := map[string]interface{}{
		"title":   title,
		"message": message,
	}
	return keyMap(streamEntity.NamedKeys(), streamVals)
}
