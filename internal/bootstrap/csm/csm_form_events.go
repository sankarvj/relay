package csm

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func events(calc, rollup string) []entity.Field {
	eventFieldID := uuid.New().String()
	eventField := entity.Field{
		Key:         eventFieldID,
		Name:        "event",
		DisplayName: "Event",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	descFieldID := uuid.New().String()
	descField := entity.Field{
		Key:         descFieldID,
		Name:        "description",
		DisplayName: "Description",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle},
	}

	countFieldID := uuid.New().String()
	countField := entity.Field{
		Key:         countFieldID,
		Name:        "count",
		DisplayName: "Count",
		DomType:     entity.DomText,
		DataType:    entity.TypeNumber,
		Meta:        map[string]string{entity.MetaKeyCalc: calc, entity.MetaKeyRollUp: rollup},
	}

	timeOfEventFieldID := uuid.New().String()
	timeOfEventField := entity.Field{
		Key:         timeOfEventFieldID,
		Name:        "time",
		DisplayName: "Time",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
		Who:         entity.WhoStartTime,
	}

	identifierFieldID := uuid.New().String()
	identifierField := entity.Field{
		Key:         identifierFieldID,
		Name:        "identifier",
		DisplayName: "Identifier",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	tagsFieldID := uuid.New().String()
	tagsField := entity.Field{
		Key:         tagsFieldID,
		Name:        "tags",
		DisplayName: "Tags",
		DataType:    entity.TypeList,
		DomType:     entity.DomMultiSelect,
		Field: &entity.Field{
			Key:      "element",
			DataType: entity.TypeString,
		},
	}

	return []entity.Field{eventField, descField, countField, timeOfEventField, identifierField, tagsField}
}
