package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func EmailConfigFields(ownerEntityID string, ownerEmailFieldKey string) []entity.Field {
	domainFieldID := uuid.New().String()
	domainField := entity.Field{
		Key:         domainFieldID,
		Name:        "domain",
		DisplayName: "Domain",
		Meta:        map[string]string{"layout": "sub-title"},
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
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

	return []entity.Field{domainField, apiKeyField, emailField, commanField, ownerField}
}

func EmailFields(emailConfigEntityID string, emailConfigOwnerFieldKey string, contactEntityID string, contactFieldKey string) []entity.Field {

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
		Meta:        map[string]string{"layout": "title"},
	}

	bodyFieldID := uuid.New().String()
	bodyField := entity.Field{
		Key:         bodyFieldID,
		Name:        "body",
		DisplayName: "Body",
		DomType:     entity.DomTextArea,
		DataType:    entity.TypeString,
	}

	return []entity.Field{fromField, toField, ccField, bccField, subjectField, bodyField}
}
