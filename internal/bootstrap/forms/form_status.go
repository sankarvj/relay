package forms

import "gitlab.com/vjsideprojects/relay/internal/entity"

func StatusFields() []entity.Field {
	verbField := entity.Field{
		Key:         entity.VerbKey, // we use this value inside the code. don't change it
		Name:        entity.Verb,
		DisplayName: "Verb (Internal field)",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "verb"},
	}

	nameField := entity.Field{
		Key:         "uuid-00-name",
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
	}

	colorField := entity.Field{
		Key:         "uuid-00-color",
		Name:        "color",
		DisplayName: "Color",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "color"},
	}

	return []entity.Field{verbField, nameField, colorField}
}

func StatusVals(verb, name, color string) map[string]interface{} {
	statusVals := map[string]interface{}{
		entity.VerbKey:  verb,
		"uuid-00-name":  name,
		"uuid-00-color": color,
	}
	return statusVals
}
