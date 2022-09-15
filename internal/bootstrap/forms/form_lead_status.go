package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func LeadStatusFields() []entity.Field {
	verbField := entity.Field{
		Key:         entity.VerbKey, // we use this value inside the code. don't change it
		Name:        entity.Verb,
		DisplayName: "Verb (Internal field)",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "verb"},
	}

	nameFieldID := uuid.New().String()
	nameField := entity.Field{
		Key:         nameFieldID,
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
	}

	colorFieldID := uuid.New().String()
	colorField := entity.Field{
		Key:         colorFieldID,
		Name:        "color",
		DisplayName: "Color",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "color"},
	}

	return []entity.Field{verbField, nameField, colorField}
}

func LeadStatusVals(statusEntity entity.Entity, verb, name, color string) map[string]interface{} {
	statusVals := map[string]interface{}{
		"name":  name,
		"color": color,
	}
	itemVals := keyMap(statusEntity.NamedKeys(), statusVals)
	itemVals[entity.VerbKey] = verb
	return itemVals
}
