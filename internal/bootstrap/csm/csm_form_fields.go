package csm

import (
	"fmt"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/reference"
)

func taskVals(actorEntity entity.Entity, desc, contactID string) map[string]interface{} {

	taskVals := make(map[string]interface{}, 0)
	namedFieldsMap := entity.NameMap(actorEntity.EasyFields())

	for name, f := range namedFieldsMap {
		switch name {
		case "desc":
			taskVals[f.Key] = desc
		case "contact":
			taskVals[f.Key] = []interface{}{contactID}
		case "reminder", "due_by":
			taskVals[f.Key] = util.FormatTimeGo(time.Now())
		}

	}
	return taskVals
}

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
		Who:         entity.WhoStatus,
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
		Meta:        map[string]string{entity.MetaKeyDisplayGex: ownerEntityKey, entity.MetaMultiChoice: "true", entity.MetaKeyLayout: entity.MetaLayoutUsers},
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
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutDate},
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

	return []entity.Field{projectNameField, planField, statusField, ownerField, contactsField, companyField, startTimeField, endTimeField, pipeField, pipeStageField}
}

func ProjVals(projectEntity entity.Entity, name string, contactID1, contactID2, flowID string) map[string]interface{} {
	projVals := map[string]interface{}{
		"project_name":         name,
		"associated_contacts":  []interface{}{contactID1, contactID2},
		"pipeline":             []interface{}{flowID},
		"pipeline_stage":       []interface{}{},
		"associated_companies": []interface{}{},
		"start_time":           util.FormatTimeGo(time.Now()),
		"end_time":             util.FormatTimeGo(time.Now().Add(10)),
	}
	return forms.KeyMap(projectEntity.NameKeyMapWrapper(), projVals)
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

func ActivitiesFields(contactEntityID, contactEntityKey, contactEntityEmailKey, companyEntityID, companyEntityKey string) []entity.Field {
	activityNameFieldID := uuid.New().String()
	activityNameField := entity.Field{
		Key:         activityNameFieldID,
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
		Who:         entity.WhoTitle,
	}

	activityDescFieldID := uuid.New().String()
	activityDescField := entity.Field{
		Key:         activityDescFieldID,
		Name:        "description",
		DisplayName: "Description",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle},
		Who:         entity.WhoDesc,
	}

	timeOfEventFieldID := uuid.New().String()
	timeOfEventField := entity.Field{
		Key:         timeOfEventFieldID,
		Name:        "time",
		DisplayName: "Time",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
		Who:         entity.WhoStartTime,
	}

	iconFieldID := uuid.New().String()
	iconField := entity.Field{
		Key:         iconFieldID,
		Name:        "icon",
		DisplayName: "Icon",
		DomType:     entity.DomImage,
		DataType:    entity.TypeString,
		Who:         entity.WhoIcon,
	}

	tagsFieldID := uuid.New().String()
	tagsField := entity.Field{
		Key:         tagsFieldID,
		Name:        "tags",
		DisplayName: "Tags",
		DataType:    entity.TypeList,
		DomType:     entity.DomMultiSelect,
		Field: &entity.Field{
			Key:      "id",
			DataType: entity.TypeString,
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
		Meta:        map[string]string{entity.MetaKeyDisplayGex: contactEntityKey, entity.MetaKeyEmailGex: contactEntityEmailKey, entity.MetaMultiChoice: "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
		Who: entity.WhoContacts,
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
		Who: entity.WhoCompanies,
	}

	return []entity.Field{activityNameField, activityDescField, timeOfEventField, tagsField, contactsField, companyField, iconField}
}

func ActivitiesVals(activityEntity entity.Entity, name, desc string, contactID1, companyID1 string) map[string]interface{} {
	actVals := map[string]interface{}{
		"name":                 name,
		"description":          desc,
		"icon":                 "🧲",
		"tags":                 []interface{}{"super"},
		"associated_contacts":  []interface{}{contactID1},
		"associated_companies": []interface{}{companyID1},
	}
	return forms.KeyMap(activityEntity.NameKeyMapWrapper(), actVals)
}

func PlanFields(contactEntityID, contactEntityKey, companyEntityID, companyEntityKey string) []entity.Field {
	planNameFieldID := uuid.New().String()
	planNameField := entity.Field{
		Key:         planNameFieldID,
		Name:        "plan_name",
		DisplayName: "Plan Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
		Who:         entity.WhoTitle,
	}

	planDescFieldID := uuid.New().String()
	planDescField := entity.Field{
		Key:         planDescFieldID,
		Name:        "activity_desc",
		DisplayName: "Description",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle},
		Who:         entity.WhoDesc,
	}

	timeOfEventFieldID := uuid.New().String()
	timeOfEventField := entity.Field{
		Key:         timeOfEventFieldID,
		Name:        "time",
		DisplayName: "Time",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
		Who:         entity.WhoStartTime,
	}

	iconFieldID := uuid.New().String()
	iconField := entity.Field{
		Key:         iconFieldID,
		Name:        "icon",
		DisplayName: "Icon",
		DomType:     entity.DomImage,
		DataType:    entity.TypeString,
		Who:         entity.WhoIcon,
	}

	reasonFieldID := uuid.New().String()
	reasonField := entity.Field{
		Key:         reasonFieldID,
		Name:        "reason",
		DisplayName: "Reason",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeList,
		Choices: []entity.Choice{
			{
				ID:           "1",
				DisplayValue: "Product not useful",
			},
			{
				ID:           "2",
				DisplayValue: "Expensive",
			},
			{
				ID:           "3",
				DisplayValue: "Technical issues",
			},
			{
				ID:           "4",
				DisplayValue: "Switching to other product",
			},
			{
				ID:           "5",
				DisplayValue: "Others",
			},
		},
		Field: &entity.Field{
			Key:      "id",
			DataType: entity.TypeString,
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
		Who: entity.WhoContacts,
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
		Who: entity.WhoCompanies,
	}

	return []entity.Field{planNameField, planDescField, timeOfEventField, reasonField, contactsField, companyField, iconField}
}

func PlanVals(planEntity entity.Entity, name, desc string, contactID1, contactID2, companyID1 string) map[string]interface{} {
	planVals := map[string]interface{}{
		"plan_name":            name,
		"plan_desc":            desc,
		"icon":                 "🔥",
		"reason":               []interface{}{util.ConvertIntToStr(randomdata.Number(1, 5))},
		"associated_contacts":  []interface{}{contactID1},
		"associated_companies": []interface{}{companyID1},
	}
	return forms.KeyMap(planEntity.NameKeyMapWrapper(), planVals)
}
