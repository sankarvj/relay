package bootstrap

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func emailConfigFields(ownerEntityID string, ownerEmailFieldKey string) []entity.Field {
	domainFieldID := uuid.New().String()
	domainField := entity.Field{
		Key:         domainFieldID,
		Name:        "domain",
		DisplayName: "Domain",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
	}

	apiKeyFieldID := uuid.New().String()
	apiKeyField := entity.Field{
		Key:         apiKeyFieldID,
		Name:        "api_key",
		DisplayName: "API Key",
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
	}

	commanFieldID := uuid.New().String()
	commanField := entity.Field{
		Key:         commanFieldID,
		Name:        "common",
		DisplayName: "",
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
		Meta:        map[string]string{"display_gex": ownerEmailFieldKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{domainField, apiKeyField, emailField, commanField, ownerField}
}

func emailFields(emailConfigEntityID string, emailConfigOwnerFieldKey string, contactEntityID string, contactFieldKey string) []entity.Field {

	configFieldID := uuid.New().String()
	configField := entity.Field{
		Key:      configFieldID,
		Name:     "config",
		DomType:  entity.DomNotApplicable,
		DataType: entity.TypeReference,
		RefID:    emailConfigEntityID,
		Meta:     map[string]string{"config": "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	fromFieldID := uuid.New().String()
	fromField := entity.Field{
		Key:         fromFieldID,
		Name:        "from",
		DisplayName: "From",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       emailConfigEntityID,
		Meta:        map[string]string{"display_gex": emailConfigOwnerFieldKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	toFieldID := uuid.New().String()
	toField := entity.Field{
		Key:         toFieldID,
		Name:        "to",
		DisplayName: "To",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       contactEntityID,
		Meta:        map[string]string{"display_gex": contactFieldKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	ccFieldID := uuid.New().String()
	ccField := entity.Field{
		Key:         ccFieldID,
		Name:        "cc",
		DisplayName: "Cc",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       emailConfigEntityID,
		Meta:        map[string]string{"display_gex": emailConfigOwnerFieldKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	bccFieldID := uuid.New().String()
	bccField := entity.Field{
		Key:         bccFieldID,
		Name:        "bcc",
		DisplayName: "Bcc",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       emailConfigEntityID,
		Meta:        map[string]string{"display_gex": emailConfigOwnerFieldKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	subjectFieldID := uuid.New().String()
	subjectField := entity.Field{
		Key:         subjectFieldID,
		Name:        "subject",
		DisplayName: "Subject",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	bodyFieldID := uuid.New().String()
	bodyField := entity.Field{
		Key:         bodyFieldID,
		Name:        "body",
		DisplayName: "Body",
		DomType:     entity.DomTextArea,
		DataType:    entity.TypeString,
	}

	return []entity.Field{configField, fromField, toField, ccField, bccField, subjectField, bodyField}
}
