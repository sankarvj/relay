package entity_test

import (
	"context"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
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
				DisplayName: "Contacts",
				AccountID:   schema.SeedAccountID,
				TeamID:      schema.SeedTeamID,
				Fields:      contactFields(),
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
		DisplayName: "First Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	emailField := entity.Field{
		DisplayName: "Email",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	mobileField := entity.Field{
		DisplayName: "Mobile Numbers",
		DataType:    entity.TypeList,
		DomType:     entity.DomMultiSelect,
		Field: &entity.Field{
			DataType: entity.TypeString,
		},
	}

	npsField := entity.Field{
		DisplayName: "NPS Score",
		DataType:    entity.TypeNumber,
	}

	lfStageField := entity.Field{
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
		DisplayName: "Status",
		DomType:     entity.DomText,
		DataType:    entity.TypeReference,
		RefID:       "refID",
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "refKey",
			Value:    "--",
		},
	}

	return []entity.Field{nameField, emailField, mobileField, npsField, lfStageField, statusField}
}

func TestSaveEmailIntegration(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	defer teardown()
	tests.SeedData(t, db)
	tests.SeedEntity(t, db)
	tests.SeedRelationShips(t, db)

	t.Log(" Save Token To The User")
	{
		t.Log("\tWhen adding token to user entity")
		{
			ctx := context.WithValue(tests.Context(), 1, &auth.Claims{
				Roles: []string{"ADMIN", "USER"},
				StandardClaims: jwt.StandardClaims{
					Subject:   "5cf37266-3473-4006-984f-9325122678b7",
					IssuedAt:  1610259806,
					ExpiresAt: 1610346206,
					NotBefore: 0,
					Audience:  "",
					Issuer:    "",
				},
			})

			_, err := entity.SaveEmailIntegration(ctx, schema.SeedAccountID, schema.SeedUserID1, "google.com", "token", "vijayasankarmail@gmail.com", db)
			if err != nil {
				t.Fatalf("\tCould not able to save token on current user - %s", err)
			}
		}
	}
}
