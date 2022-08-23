package forms

import "gitlab.com/vjsideprojects/relay/internal/entity"

func StreamFields() []entity.Field {
	titleField := entity.Field{
		Key:         "uuid-00-title",
		Name:        "title",
		DisplayName: "Title",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	messageField := entity.Field{
		Key:         "uuid-00-message",
		Name:        "message",
		DisplayName: "Message",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle},
	}

	return []entity.Field{titleField, messageField}
}

func StreamVals(title, message, file string) map[string]interface{} {
	streamVals := map[string]interface{}{
		"uuid-00-title":   title,
		"uuid-00-message": message,
	}
	return streamVals
}
