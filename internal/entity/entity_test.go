package entity_test

import (
	"encoding/json"
	"testing"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/tests"
)

func TestDataEntity(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	defer teardown()
	tests.SeedData(t, db)

	t.Log(" Given the need to create and update an data entity.")
	{

		t.Log("\tPrepare status reference entity")
		{
			ne := entity.NewEntity{
				ID:          "00000000-0000-0000-0002-000000000000",
				DisplayName: "Status",
				AccountID:   schema.SeedAccountID,
				TeamID:      schema.SeedTeamID,
				Fields:      dummyFields(),
			}
			_, err := entity.Create(tests.Context(), db, ne, time.Now())
			if err != nil {
				t.Fatalf("\tShould be able to create an status entity - %s", err)
			}
		}

		t.Log("\tPrepare country reference entity")
		{
			ne := entity.NewEntity{
				ID:          "00000000-0000-0000-0003-000000000000",
				DisplayName: "Countries",
				AccountID:   schema.SeedAccountID,
				TeamID:      schema.SeedTeamID,
				Fields:      dummyFields(),
			}
			_, err := entity.Create(tests.Context(), db, ne, time.Now())
			if err != nil {
				t.Fatalf("\tShould be able to prepare an entity - %s", err)
			}
		}

		t.Log("\tWhen adding the data entity")
		{
			ne := entity.NewEntity{
				ID:          "00000000-0000-0000-0001-000000000000",
				DisplayName: "Contacts",
				AccountID:   schema.SeedAccountID,
				TeamID:      schema.SeedTeamID,
				Fields:      contactFields(""),
			}
			_, err := entity.Create(tests.Context(), db, ne, time.Now())
			if err != nil {
				t.Fatalf("\tShould be able to create an entity - %s", err)
			}
		}

		t.Log("\tWhen updating the data entity")
		{
			input, err := json.Marshal(contactFields("00000000-0000-0000-0002-000000000000"))
			if err != nil {
				t.Fatalf("\tShould be able to marshal the fields - %s", err)
			}
			err = entity.Update(tests.Context(), db, schema.SeedAccountID, "00000000-0000-0000-0001-000000000000", string(input), time.Now())
			if err != nil {
				t.Fatalf("\tShould be able to update an entity - %s", err)
			}
		}

		t.Log("\tRetriving the relationships added by the above tests")
		{

			relationships, err := relationship.Relationships(tests.Context(), db, schema.SeedAccountID, "00000000-0000-0000-0001-000000000000")
			if err != nil || len(relationships) != 2 {
				t.Fatalf("\tShould be able to get two relations - %s", err)
			}
		}
	}
}

func contactFields(statusRefID string) []entity.Field {
	nameField := entity.Field{
		Key:         "00000000-0000-0000-0000-000000000001",
		DisplayName: "First Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	emailField := entity.Field{
		Key:         "00000000-0000-0000-0000-000000000002",
		DisplayName: "Email",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	mobileField := entity.Field{
		Key:         "00000000-0000-0000-0000-000000000003",
		DisplayName: "Mobile Numbers",
		DataType:    entity.TypeList,
		DomType:     entity.DomMultiSelect,
		Field: &entity.Field{
			DataType: entity.TypeString,
		},
	}

	npsField := entity.Field{
		Key:         "00000000-0000-0000-0000-000000000004",
		DisplayName: "NPS Score",
		DataType:    entity.TypeNumber,
	}

	lfStageField := entity.Field{
		Key:         "00000000-0000-0000-0000-000000000005",
		DisplayName: "Lifecycle Stage",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeString,
		Choices: []entity.Choice{
			{ID: "uuid_lead", DisplayValue: "lead"},
			{ID: "uuid_contact", DisplayValue: "contact"},
			{ID: "uuid_won", DisplayValue: "won"},
		},
	}

	statusField := entity.Field{
		Key:         "00000000-0000-0000-0000-000000000006",
		DisplayName: "Status",
		DomType:     entity.DomText,
		DataType:    entity.TypeReference,
		RefID:       statusRefID,
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	countryField := entity.Field{
		Key:         "00000000-0000-0000-0000-000000000007",
		DisplayName: "Country",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       "00000000-0000-0000-0003-000000000000",
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{nameField, emailField, mobileField, npsField, lfStageField, statusField, countryField}
}

func dummyFields() []entity.Field {
	nameField := entity.Field{
		Key:         "00000000-0000-0000-0000-000000000011",
		DisplayName: "Dummy",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	return []entity.Field{nameField}
}
