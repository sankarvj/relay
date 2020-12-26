package entity_test

import (
	"testing"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/tests"
)

func TestDataEntity(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	defer teardown()
	tests.SeedData(t, db)

	t.Log(" Given the need to create an data entity.")
	{
		t.Log("\tWhen adding the data entity")
		{
			ne := entity.NewEntity{
				Name:      "Contacts",
				AccountID: schema.SeedAccountID,
				TeamID:    schema.SeedTeamID1,
				Fields:    contactFields(),
			}
			_, err := entity.Create(tests.Context(), db, ne, time.Now())
			if err != nil {
				t.Fatalf("\tShould not be able to create an entity - %s", err)
			}
		}
	}
}

func contactFields() []entity.Field {
	nameField := entity.Field{
		Name:     "First Name",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
	}

	emailField := entity.Field{
		Name:     "Email",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
	}

	mobileField := entity.Field{
		Name:     "Mobile Numbers",
		DataType: entity.TypeList,
		DomType:  entity.DomMultiSelect,
		Field: &entity.Field{
			DataType: entity.TypeString,
		},
	}

	npsField := entity.Field{
		Name:     "NPS Score",
		DataType: entity.TypeNumber,
	}

	lfStageField := entity.Field{
		Name:     "Lifecycle Stage",
		DomType:  entity.DomSelect,
		DataType: entity.TypeString,
		Choices: []entity.Choice{
			{ID: "uuid_lead", DisplayValue: "lead"},
			{ID: "uuid_contact", DisplayValue: "contact"},
			{ID: "uuid_won", DisplayValue: "won"},
		},
	}

	statusField := entity.Field{
		Name:     "Status",
		DomType:  entity.DomText,
		DataType: entity.TypeReference,
		RefID:    "refID",
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "refKey",
			Value:    "--",
		},
	}

	return []entity.Field{nameField, emailField, mobileField, npsField, lfStageField, statusField}
}
