package csm

import (
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/reference"
)

func ProjectFields(statusEntityID, statusEntityKey, ownerEntityID, ownerEntityKey, contactEntityID, contactEntityKey, companyEntityID, companyEntityKey string, flowEntityID, nodeEntityID, nodeKey string) []entity.Field {
	projectNameFieldID := uuid.New().String()
	projectNameField := entity.Field{
		Key:         projectNameFieldID,
		Name:        "project_name",
		DisplayName: "Project Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
	}

	planFieldID := uuid.New().String()
	planField := entity.Field{
		Key:         planFieldID,
		Name:        "plan",
		DisplayName: "Plan",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	statusFieldID := uuid.New().String()
	statusField := entity.Field{
		Key:         statusFieldID,
		Name:        "status",
		DisplayName: "Status",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       statusEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: statusEntityKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	ownerFieldID := uuid.New().String()
	ownerField := entity.Field{
		Key:         ownerFieldID,
		Name:        "owner",
		DisplayName: "Project Owner",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       ownerEntityID,
		RefType:     entity.RefTypeStraight,
		Who:         entity.WhoAssignee,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: ownerEntityKey, entity.MetaMultiChoice: "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	contactsFieldID := uuid.New().String()
	contactsField := entity.Field{
		Key:         contactsFieldID,
		Name:        "associated_contacts",
		DisplayName: "Associated Contacts",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       contactEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: contactEntityKey, entity.MetaMultiChoice: "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	companyFieldID := uuid.New().String()
	companyField := entity.Field{
		Key:         companyFieldID,
		Name:        "associated_companies",
		DisplayName: "Associated Companies",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       companyEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: companyEntityKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	pipeFieldID := uuid.New().String()
	pipeField := entity.Field{
		Key:         pipeFieldID,
		Name:        "pipeline",
		DisplayName: "Pipeline",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       flowEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyFlow: "true", entity.MetaMultiChoice: "false"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	pipeStageFieldID := uuid.New().String()
	pipeStageField := entity.Field{
		Key:         pipeStageFieldID,
		Name:        "pipeline_stage",
		DisplayName: "Pipeline Stage",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       nodeEntityID,
		RefType:     entity.RefTypeStraight,
		Dependent: &entity.Dependent{
			ParentKey:   pipeField.Key,
			Expressions: []string{""}, // empty means positive
			Actions:     []string{fmt.Sprintf("{{{%s.%s}}}", reference.ActionFilter, reference.ByFlow)},
		},
		Meta: map[string]string{entity.MetaKeyDisplayGex: nodeKey, entity.MetaKeyNode: "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      pipeField.Key,
			Value:    "--",
		},
	}

	return []entity.Field{projectNameField, planField, statusField, ownerField, contactsField, companyField, pipeField, pipeStageField}
}

func ProjVals(projectEntity entity.Entity, name string, amount int, contactID1, contactID2, flowID string) map[string]interface{} {
	projVals := map[string]interface{}{
		"project_name":         name,
		"associated_contacts":  []interface{}{contactID1, contactID2},
		"pipeline":             []interface{}{flowID},
		"pipeline_stage":       []interface{}{},
		"associated_companies": []interface{}{},
	}
	return forms.KeyMap(projectEntity.NamedKeys(), projVals)
}

func MeetingFields(contactEntityID, companyEntityID, projectEntityID string, contactEntityEmailFieldID, contactEntityFirstNameFieldID, companyEntityNameFieldID, projectEntityNameFieldID string) []entity.Field {

	titleFieldID := uuid.New().String()
	titleField := entity.Field{
		Key:         titleFieldID,
		Name:        "cal_title",
		DisplayName: "Title",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
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
		Meta:        map[string]string{entity.MetaKeyRow: "true"},
		Who:         entity.WhoEndTime,
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

	projectFieldID := uuid.New().String()
	projectField := entity.Field{
		Key:         projectFieldID,
		Name:        "project",
		DisplayName: "Associated Project",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       projectEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: projectEntityNameFieldID, entity.MetaKeyRow: "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{titleField, summaryField, attendessField, startTimeField, endTimeField, timezoneField, createdAtField, updatedAtField, contactField, companyField, projectField}
}
