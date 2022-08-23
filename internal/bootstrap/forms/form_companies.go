package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func CompanyFields(ownerEntityID string, ownerEntityKey string) []entity.Field {
	nameFieldID := uuid.New().String()
	nameField := entity.Field{
		Key:         nameFieldID,
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
	}

	websiteFieldID := uuid.New().String()
	websiteField := entity.Field{
		Key:         websiteFieldID,
		Name:        "website",
		DisplayName: "Domain",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	cityFieldID := uuid.New().String()
	cityField := entity.Field{
		Key:         cityFieldID,
		Name:        "city",
		DisplayName: "City",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	stateFieldID := uuid.New().String()
	stateField := entity.Field{
		Key:         stateFieldID,
		Name:        "state",
		DisplayName: "State",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	annualRevenueFieldID := uuid.New().String()
	annualRevenueField := entity.Field{
		Key:         annualRevenueFieldID,
		Name:        "revenue",
		DisplayName: "Annual Revenue",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	countryFieldID := uuid.New().String()
	countryField := entity.Field{
		Key:         countryFieldID,
		Name:        "country",
		DisplayName: "Country",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	employeesCountFieldID := uuid.New().String()
	employeesCountField := entity.Field{
		Key:         employeesCountFieldID,
		Name:        "employees_count",
		DisplayName: "Employees Count",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	ownerFieldID := uuid.New().String()
	ownerField := entity.Field{
		Key:         ownerFieldID,
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

func CompanyVals(companyEntity entity.Entity, name, website string) map[string]interface{} {
	namedVals := map[string]interface{}{
		"name":            name,
		"website":         website,
		"city":            "san francisco",
		"state":           "california",
		"country":         "USA",
		"employees_count": 1000,
		"revenue":         "2000",
		"owner":           []interface{}{},
	}

	return keyMap(companyEntity.NamedKeys(), namedVals)
}
