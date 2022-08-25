package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func MeetingFields(contactEntityID, companyEntityID, contactEntityEmailFieldID, contactEntityFirstNameFieldID, companyEntityNameFieldID string) []entity.Field {

	titleFieldID := uuid.New().String()
	titleField := entity.Field{
		Key:         titleFieldID,
		Name:        "cal_title",
		DisplayName: "Title",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	summaryFieldID := uuid.New().String()
	summaryField := entity.Field{
		Key:         summaryFieldID,
		Name:        "summary",
		DisplayName: "Summary",
		DomType:     entity.DomTextArea,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle},
	}

	attendessFieldID := uuid.New().String()
	attendessField := entity.Field{
		Key:         attendessFieldID,
		Name:        "attendess",
		DisplayName: "Attendess",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       contactEntityID,
		Who:         entity.WhoAssignee,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: contactEntityEmailFieldID, entity.MetaKeyLayout: entity.MetaLayoutUsers},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	startTimeFieldID := uuid.New().String()
	startTimeField := entity.Field{
		Key:         startTimeFieldID,
		Name:        "start_time",
		DisplayName: "Start Time",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
		Who:         entity.WhoStartTime,
	}

	endTimeFieldID := uuid.New().String()
	endTimeField := entity.Field{
		Key:         endTimeFieldID,
		Name:        "end_time",
		DisplayName: "End Time",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
		Who:         entity.WhoEndTime,
		Meta:        map[string]string{entity.MetaKeyRow: "true"},
	}

	timezoneFieldID := uuid.New().String()
	timezoneField := entity.Field{
		Key:         timezoneFieldID,
		Name:        "timezone",
		DisplayName: "Timezone",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyHidden: "true"},
	}

	createdAtFieldID := uuid.New().String()
	createdAtField := entity.Field{
		Key:         createdAtFieldID,
		Name:        "created_at",
		DisplayName: "Created At",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
		Meta:        map[string]string{entity.MetaKeyHidden: "true"},
	}

	updatedAtFieldID := uuid.New().String()
	updatedAtField := entity.Field{
		Key:         updatedAtFieldID,
		Name:        "updated_at",
		DisplayName: "Updated At",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
		Meta:        map[string]string{entity.MetaKeyHidden: "true"},
	}

	contactFieldID := uuid.New().String()
	contactField := entity.Field{
		Key:         contactFieldID,
		Name:        "contact",
		DisplayName: "Associated Contact",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       contactEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: contactEntityFirstNameFieldID},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	companyFieldID := uuid.New().String()
	companyField := entity.Field{
		Key:         companyFieldID,
		Name:        "company",
		DisplayName: "Associated Company",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       companyEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: companyEntityNameFieldID, entity.MetaKeyRow: "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{titleField, summaryField, attendessField, startTimeField, endTimeField, timezoneField, createdAtField, updatedAtField, contactField, companyField}
}
