package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func VisitorInvitationFields() []entity.Field {

	emailFieldID := uuid.New().String()
	emailField := entity.Field{
		Key:         emailFieldID,
		Name:        "email",
		DisplayName: "To Email",
		DomType:     entity.DomEmailSelector,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyHTML: "true"},
	}

	roleFieldID := uuid.New().String()
	roleField := entity.Field{
		Key:         roleFieldID,
		Name:        "role",
		DisplayName: "Role",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeList,
		Meta:        map[string]string{entity.MetaMultiChoice: "false"},
		Choices: []entity.Choice{
			{
				ID:           "ADMIN",
				DisplayValue: "ADMIN",
			},
			{
				ID:           "MEMBER",
				DisplayValue: "MEMBER",
			},
			{
				ID:           "USER",
				DisplayValue: "USER",
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

	bodyFieldID := uuid.New().String()
	bodyField := entity.Field{
		Key:         bodyFieldID,
		Name:        "body",
		DisplayName: "Message",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyHTML: "true"},
	}

	return []entity.Field{emailField, roleField, bodyField}
}
