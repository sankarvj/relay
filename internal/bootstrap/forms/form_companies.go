package forms

import (
	"github.com/Pallinder/go-randomdata"
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
		DataType:    entity.TypeNumber,
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
		DataType:    entity.TypeNumber,
	}

	healthFieldID := uuid.New().String()
	healthField := entity.Field{
		Key:         healthFieldID,
		Name:        "health",
		DisplayName: "Health",
		DomType:     entity.DomText,
		DataType:    entity.TypeNumber,
		Choices: []entity.Choice{
			{
				ID:           "1",
				DisplayValue: "Very Poor",
			},
			{
				ID:           "2",
				DisplayValue: "Poor",
			},
			{
				ID:           "3",
				DisplayValue: "Moderate",
			},
			{
				ID:           "4",
				DisplayValue: "Good",
			},
			{
				ID:           "5",
				DisplayValue: "Cool",
			},
		},
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

	return []entity.Field{nameField, websiteField, cityField, stateField, ownerField, annualRevenueField, countryField, employeesCountField, healthField}
}

func CompanyVals(companyEntity entity.Entity, name, website string) map[string]interface{} {
	namedVals := map[string]interface{}{
		"name":            name,
		"website":         website,
		"city":            randomdata.City(),
		"state":           randomdata.State(randomdata.Large),
		"country":         "USA",
		"employees_count": randomdata.Number(200),
		"revenue":         randomdata.Number(3000),
		"health":          randomdata.Number(1, 5),
		"owner":           []interface{}{},
	}

	return keyMap(companyEntity.NamedKeys(), namedVals)
}
