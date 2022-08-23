package forms

import "gitlab.com/vjsideprojects/relay/internal/entity"

func TicketFields(contactEntityID, contactEntityKey, companyEntityID, companyEntityKey, statusEntityID, statusEntityKey string) []entity.Field {
	nameField := entity.Field{
		Key:         "uuid-00-subject",
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
	}

	statusField := entity.Field{
		Key:         "uuid-00-status",
		Name:        "status",
		DisplayName: "Status",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       statusEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: statusEntityKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	contactField := entity.Field{
		Key:         "uuid-00-contact",
		Name:        "contact",
		DisplayName: "Associated Contacts",
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
		DisplayName: "Associated Companies",
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

	return []entity.Field{nameField, statusField, contactField, companyField}
}
