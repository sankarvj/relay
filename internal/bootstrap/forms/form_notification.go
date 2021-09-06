package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func NotificationFields() []entity.Field {
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

	return []entity.Field{subjectField, bodyField, typeField, timeField, accountField, entityField, itemField}
}
