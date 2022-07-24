package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func EmailConfigFields(ownerEntityID string, ownerEmailFieldKey string) []entity.Field {

	accountFieldID := uuid.New().String()
	accountField := entity.Field{
		Key:         accountFieldID,
		Name:        "account_id",
		DisplayName: "",
		Meta:        map[string]string{entity.MetaKeyConfig: "true"},
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
	}

	teamFieldID := uuid.New().String()
	teamField := entity.Field{
		Key:         teamFieldID,
		Name:        "team_id",
		DisplayName: "",
		Meta:        map[string]string{entity.MetaKeyConfig: "true"},
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
	}

	domainFieldID := uuid.New().String()
	domainField := entity.Field{
		Key:         domainFieldID,
		Name:        "domain",
		DisplayName: "Domain",
		Meta:        map[string]string{entity.MetaKeyLayout: "sub-title"},
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
	}

	apiKeyFieldID := uuid.New().String()
	apiKeyField := entity.Field{
		Key:         apiKeyFieldID,
		Name:        "api_key",
		DisplayName: "API Key",
		Meta:        map[string]string{entity.MetaKeyConfig: "true"},
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
		Meta:        map[string]string{entity.MetaKeyLayout: "super-title"}, //super-title overwrites title if exists
	}

	commanFieldID := uuid.New().String()
	commanField := entity.Field{
		Key:         commanFieldID,
		Name:        "common",
		DisplayName: "",
		Meta:        map[string]string{entity.MetaKeyConfig: "true"},
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
	}

	historyFieldID := uuid.New().String()
	historyField := entity.Field{
		Key:         historyFieldID,
		Name:        "history_id",
		DisplayName: "",
		Meta:        map[string]string{entity.MetaKeyConfig: "true"},
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
		Meta:        map[string]string{entity.MetaKeyDisplayGex: ownerEmailFieldKey, entity.MetaKeyLayout: "title"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{accountField, teamField, domainField, apiKeyField, emailField, commanField, ownerField, historyField}
}

func EmailFields(emailConfigEntityID string, emailConfigOwnerFieldKey string) []entity.Field {

	messageFieldID := uuid.New().String()
	messageField := entity.Field{
		Key:         messageFieldID,
		Name:        "message_id",
		DisplayName: "Message ID",
		DataType:    entity.TypeString,
		DomType:     entity.DomNotApplicable,
	}

	messageSentFieldID := uuid.New().String()
	messageSentField := entity.Field{
		Key:         messageSentFieldID,
		Name:        "message_sent",
		DisplayName: "Message Sent",
		DataType:    entity.TypeString,
		DomType:     entity.DomNotApplicable,
	}

	fromFieldID := uuid.New().String()
	fromField := entity.Field{
		Key:         fromFieldID,
		Name:        "from",
		DisplayName: "From",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       emailConfigEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: emailConfigOwnerFieldKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	receivingfromFieldID := uuid.New().String()
	receivingfromField := entity.Field{
		Key:         receivingfromFieldID,
		Name:        "rfrom",
		DisplayName: "Receving From",
		DataType:    entity.TypeList,
		DomType:     entity.DomMultiSelect,
		Meta:        map[string]string{entity.MetaKeyHidden: "true"},
		Field: &entity.Field{
			Key:      "element",
			DataType: entity.TypeString,
		},
	}

	toFieldID := uuid.New().String()
	toField := entity.Field{
		Key:         toFieldID,
		Name:        "to",
		DisplayName: "To",
		DataType:    entity.TypeList,
		DomType:     entity.DomMultiSelect,
		Field: &entity.Field{
			Key:      "element",
			DataType: entity.TypeString,
		},
	}

	ccFieldID := uuid.New().String()
	ccField := entity.Field{
		Key:         ccFieldID,
		Name:        "cc",
		DisplayName: "Cc",
		DataType:    entity.TypeList,
		DomType:     entity.DomMultiSelect,
		Field: &entity.Field{
			Key:      "element",
			DataType: entity.TypeString,
		},
	}

	bccFieldID := uuid.New().String()
	bccField := entity.Field{
		Key:         bccFieldID,
		Name:        "bcc",
		DisplayName: "Bcc",
		DataType:    entity.TypeList,
		DomType:     entity.DomMultiSelect,
		Field: &entity.Field{
			Key:      "element",
			DataType: entity.TypeString,
		},
	}

	subjectFieldID := uuid.New().String()
	subjectField := entity.Field{
		Key:         subjectFieldID,
		Name:        "subject",
		DisplayName: "Subject",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title", entity.MetaKeyHTML: "true"},
	}

	bodyFieldID := uuid.New().String()
	bodyField := entity.Field{
		Key:         bodyFieldID,
		Name:        "body",
		DisplayName: "Body",
		DomType:     entity.DomTextArea,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyHTML: "true"},
	}

	// params: b.ContactEntity.ID, b.CompanyEntity.ID, b.ContactEntity.Key("first_name"), b.CompanyEntity.Key("name")
	// contactFieldID := uuid.New().String()
	// contactField := entity.Field{
	// 	Key:         contactFieldID,
	// 	Name:        "contacts",
	// 	DisplayName: "Associated Contacts",
	// 	DomType:     entity.DomAutoComplete,
	// 	DataType:    entity.TypeReference,
	// 	RefID:       contactEntityID,
	// 	Meta:        map[string]string{entity.MetaKeyDisplayGex: contactNameKey},
	// 	Field: &entity.Field{
	// 		DataType: entity.TypeString,
	// 		Key:      "id",
	// 		Value:    "--",
	// 	},
	// }

	// companyFieldID := uuid.New().String()
	// companyField := entity.Field{
	// 	Key:         companyFieldID,
	// 	Name:        "companies",
	// 	DisplayName: "Associated Companies",
	// 	DomType:     entity.DomAutoComplete,
	// 	DataType:    entity.TypeReference,
	// 	RefID:       companyEntityID,
	// 	Meta:        map[string]string{entity.MetaKeyDisplayGex: companyNameKey},
	// 	Field: &entity.Field{
	// 		DataType: entity.TypeString,
	// 		Key:      "id",
	// 		Value:    "--",
	// 	},
	// }

	return []entity.Field{messageField, messageSentField, receivingfromField, fromField, toField, ccField, bccField, subjectField, bodyField}
}
