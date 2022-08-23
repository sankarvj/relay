package forms

import "gitlab.com/vjsideprojects/relay/internal/entity"

func NoteFields(contactEntityID, contactEntityKey, companyEntityID, companyEntityKey string) []entity.Field {
	descField := entity.Field{
		Key:         "uuid-00-desc",
		Name:        "desc",
		DisplayName: "Notes",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	contactField := entity.Field{
		Key:         "uuid-00-contact",
		Name:        "contact",
		DisplayName: "Associated To",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       contactEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: contactEntityKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	companyField := entity.Field{
		Key:         "uuid-00-company",
		Name:        "company",
		DisplayName: "Associated To",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       companyEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: companyEntityKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{descField, contactField, companyField}
}
