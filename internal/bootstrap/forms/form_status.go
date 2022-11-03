package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func StatusFields() []entity.Field {
	verbFieldID := uuid.New().String()
	verbField := entity.Field{
		Key:         verbFieldID, // we use this value inside the code. don't change it
		Name:        "verb",
		DisplayName: "Verb (Internal field)",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutVerb},
		Who:         entity.WhoVerb,
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

	return []entity.Field{verbField, nameField, colorField}
}

func StatusVals(statusEntity entity.Entity, verb, name, color string) map[string]interface{} {
	statusVals := map[string]interface{}{
		"verb":  verb,
		"name":  name,
		"color": color,
	}
	itemVals := keyMap(statusEntity.NameKeyMapWrapper(), statusVals)
	return itemVals
}
