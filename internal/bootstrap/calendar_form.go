package bootstrap

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func calendarFields(ownerEntityID string, ownerEmailFieldKey string) []entity.Field {
	domainFieldID := uuid.New().String()
	idField := entity.Field{
		Key:         domainFieldID,
		Name:        "id",
		DisplayName: "Calendar ID",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "sub-title"},
	}

	apiKeyFieldID := uuid.New().String()
	apiKeyField := entity.Field{
		Key:         apiKeyFieldID,
		Name:        "api_key",
		DisplayName: "API Key",
		Meta:        map[string]string{"config": "true"},
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
	}

	emailFieldID := uuid.New().String()
	emailField := entity.Field{
		Key:         emailFieldID,
		Name:        "email",
		DisplayName: "E-Mail",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "super-title"}, //super-title overwrites title if exists
	}

	commanFieldID := uuid.New().String()
	commanField := entity.Field{
		Key:         commanFieldID,
		Name:        "common",
		DisplayName: "",
		Meta:        map[string]string{"config": "true"},
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
	}

	ownerFieldID := uuid.New().String()
	ownerField := entity.Field{
		Key:         ownerFieldID,
		Name:        "owner",
		DisplayName: "Associated To",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       ownerEntityID,
		Meta:        map[string]string{"display_gex": ownerEmailFieldKey, "layout": "title"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	syncTokenFieldID := uuid.New().String()
	syncTokenField := entity.Field{
		Key:         syncTokenFieldID,
		Name:        "sync_token",
		DisplayName: "Sync Token",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	syncedAtFieldID := uuid.New().String()
	syncedAtField := entity.Field{
		Key:         syncedAtFieldID,
		Name:        "synced_at",
		DisplayName: "Last Synced",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
	}

	retriesFieldID := uuid.New().String()
	retriesField := entity.Field{
		Key:         retriesFieldID,
		Name:        "retries",
		DisplayName: "Retries",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	return []entity.Field{idField, apiKeyField, emailField, commanField, ownerField, syncTokenField, syncedAtField, retriesField}
}
