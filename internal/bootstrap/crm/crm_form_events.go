package crm

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func pageViewEventEntityFields() []entity.Field {
	urlFieldID := uuid.New().String()
	urlField := entity.Field{
		Key:         urlFieldID,
		Name:        "url",
		DisplayName: "URL",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "title"},
	}

	vistsCountFieldID := uuid.New().String()
	vistsCountField := entity.Field{
		Key:         vistsCountFieldID,
		Name:        "visits",
		DisplayName: "Vists",
		DomType:     entity.DomText,
		DataType:    entity.TypeNumber,
		Meta:        map[string]string{"layout": "footer"},
	}

	linkFieldID := uuid.New().String()
	linkField := entity.Field{
		Key:         linkFieldID,
		Name:        "link",
		DisplayName: "Link",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "link"},
	}

	return []entity.Field{urlField, vistsCountField, linkField}
}

func activityEventEntityFields() []entity.Field {

	activityNameFieldID := uuid.New().String()
	activityNameField := entity.Field{
		Key:         activityNameFieldID,
		Name:        "activity-name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "title"},
	}

	activityActionFieldID := uuid.New().String()
	activityActionField := entity.Field{
		Key:         activityActionFieldID,
		Name:        "activity-action",
		DisplayName: "Action",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "footer"},
	}

	linkFieldID := uuid.New().String()
	linkField := entity.Field{
		Key:         linkFieldID,
		Name:        "activity-link",
		DisplayName: "Link",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "link"},
	}

	return []entity.Field{activityNameField, activityActionField, linkField}
}

func propertyChangeEventEntityFields() []entity.Field {
	propertyNameFieldID := uuid.New().String()
	propertyNameField := entity.Field{
		Key:         propertyNameFieldID,
		Name:        "property-name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "title"},
	}

	propertyValueFieldID := uuid.New().String()
	propertyValueField := entity.Field{
		Key:         propertyValueFieldID,
		Name:        "property-value",
		DisplayName: "Value",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "footer"},
	}

	linkFieldID := uuid.New().String()
	linkField := entity.Field{
		Key:         linkFieldID,
		Name:        "link",
		DisplayName: "Link",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "link"},
	}

	return []entity.Field{propertyNameField, propertyValueField, linkField}
}
