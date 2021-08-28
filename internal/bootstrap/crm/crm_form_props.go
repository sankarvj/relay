package crm

import (
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func pageViewEventEntityFields() []entity.Field {

	urlField := entity.Field{
		Key:         "uuid-00-url",
		Name:        "url",
		DisplayName: "URL",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "title"},
	}

	vistsCountField := entity.Field{
		Key:         "uuid-00-visits",
		Name:        "visits",
		DisplayName: "Vists",
		DomType:     entity.DomText,
		DataType:    entity.TypeNumber,
		Meta:        map[string]string{"layout": "footer"},
	}

	linkField := entity.Field{
		Key:         "uuid-00-link-key",
		Name:        "link",
		DisplayName: "Link",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "link"},
	}

	return []entity.Field{urlField, vistsCountField, linkField}
}

func activityEventEntityFields() []entity.Field {

	activityNameField := entity.Field{
		Key:         "uuid-00-activity-name",
		Name:        "activity-name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "title"},
	}

	activityActionField := entity.Field{
		Key:         "uuid-00-activity-action",
		Name:        "activity-action",
		DisplayName: "Action",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "footer"},
	}

	linkField := entity.Field{
		Key:         "uuid-00-link-key",
		Name:        "activity-link",
		DisplayName: "Link",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "link"},
	}

	return []entity.Field{activityNameField, activityActionField, linkField}
}

func propertyChangeEventEntityFields() []entity.Field {
	propertyNameField := entity.Field{
		Key:         "uuid-00-property-name",
		Name:        "property-name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "title"},
	}

	propertyValueField := entity.Field{
		Key:         "uuid-00-property-value",
		Name:        "property-value",
		DisplayName: "Value",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "footer"},
	}

	linkField := entity.Field{
		Key:         "uuid-00-link-key",
		Name:        "link",
		DisplayName: "Link",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "link"},
	}

	return []entity.Field{propertyNameField, propertyValueField, linkField}
}
