package forms

import "gitlab.com/vjsideprojects/relay/internal/entity"

func CompanyFields(ownerEntityID string, ownerEntityKey string) []entity.Field {
	nameField := entity.Field{
		Key:         "uuid-00-name",
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
	}

	websiteField := entity.Field{
		Key:         "uuid-00-website",
		Name:        "website",
		DisplayName: "Domain",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	cityField := entity.Field{
		Key:         "uuid-00-city",
		Name:        "city",
		DisplayName: "City",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	stateField := entity.Field{
		Key:         "uuid-00-state",
		Name:        "state",
		DisplayName: "State",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	annualRevenueField := entity.Field{
		Key:         "uuid-00-revenue",
		Name:        "revenue",
		DisplayName: "Annual Revenue",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	countryField := entity.Field{
		Key:         "uuid-00-country",
		Name:        "country",
		DisplayName: "Country",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	employeesCountField := entity.Field{
		Key:         "uuid-00-employees-count",
		Name:        "employees_count",
		DisplayName: "Employees Count",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	ownerField := entity.Field{
		Key:         "uuid-00-owner",
		Name:        "owner",
		DisplayName: "Company Owner",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       ownerEntityID,
		RefType:     entity.RefTypeStraight,
		Who:         entity.WhoAssignee,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: ownerEntityKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{nameField, websiteField, cityField, stateField, ownerField, annualRevenueField, countryField, employeesCountField}
}

func CompanyVals(name, website string) map[string]interface{} {
	companyVals := map[string]interface{}{
		"uuid-00-name":            name,
		"uuid-00-website":         website,
		"uuid-00-city":            "san francisco",
		"uuid-00-state":           "california",
		"uuid-00-country":         "USA",
		"uuid-00-employees-count": 1000,
		"uuid-00-revenue":         "2000",
		"uuid-00-owner":           []interface{}{},
	}
	return companyVals
}
