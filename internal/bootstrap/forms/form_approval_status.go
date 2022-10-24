package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func ApprovalStatusFields() []entity.Field {
	verbFieldID := uuid.New().String()
	verbField := entity.Field{
		Key:         verbFieldID,
		Name:        "verb",
		DisplayName: "Verb (Internal field)",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutVerb},
		Who:         entity.WhoVerb,
	}

	identifierFieldID := uuid.New().String()
	identifierField := entity.Field{
		Key:         identifierFieldID,
		Name:        "identifier",
		DisplayName: "Identifier (Internal field)",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
		Who:         entity.WhoIdentifier,
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

	colorFieldID := uuid.New().String()
	colorField := entity.Field{
		Key:         colorFieldID,
		Name:        "color",
		DisplayName: "Color",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutColor},
		Who:         entity.WhoColor,
	}

	return []entity.Field{verbField, identifierField, nameField, colorField}
}

func ApprovalStatusVals(approvalStatusEntity entity.Entity, verb, name, identifier, color string) map[string]interface{} {
	statusVals := map[string]interface{}{
		"verb":       verb,
		"name":       name,
		"identifier": identifier,
		"color":      color,
	}
	itemVals := keyMap(approvalStatusEntity.NamedKeys(), statusVals)
	return itemVals
}
