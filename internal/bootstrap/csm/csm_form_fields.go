package csm

import (
	"fmt"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"gitlab.com/vjsideprojects/relay/internal/schema"
)

func ProjectFields(statusEntityID, ownerEntityID, ownerEntityKey, contactEntityID, companyEntityID string, flowEntityID, nodeEntityID string) []entity.Field {
	projectNameField := entity.Field{
		Key:         "uuid-00-project-name",
		Name:        "project_name",
		DisplayName: "Project Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
	}

	planField := entity.Field{
		Key:         schema.SeedFieldPlanKey,
		Name:        "plan",
		DisplayName: "Plan",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	statusField := entity.Field{
		Key:         "uuid-00-status",
		Name:        "status",
		DisplayName: "Status",
		DomType:     entity.DomAutoSelect,
		DataType:    entity.TypeReference,
		RefID:       statusEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: "uuid-00-name"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	ownerField := entity.Field{
		Key:         "uuid-00-proj-owner",
		Name:        "owner",
		DisplayName: "Project Owner",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       ownerEntityID,
		RefType:     entity.RefTypeStraight,
		Who:         entity.WhoAssignee,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: ownerEntityKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	contactsField := entity.Field{
		Key:         "uuid-00-contacts",
		Name:        "contact",
		DisplayName: "Associated Contacts",
		DomType:     entity.DomMultiSelect,
		DataType:    entity.TypeReference,
		RefID:       contactEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: "uuid-00-fname"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	companyField := entity.Field{
		Key:         "uuid-00-company",
		Name:        "company",
		DisplayName: "Associated Companies",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       companyEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: "uuid-00-name"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	pipeField := entity.Field{
		Key:         "uuid-00-pipe",
		Name:        "pipeline",
		DisplayName: "Pipeline",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       flowEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyFlow: "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	pipeStageField := entity.Field{
		Key:         "uuid-00-pipe-stage",
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
		Meta: map[string]string{entity.MetaKeyDisplayGex: "uuid-00-fname", "node": "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      pipeField.Key,
			Value:    "--",
		},
	}

	return []entity.Field{projectNameField, planField, statusField, ownerField, contactsField, companyField, pipeField, pipeStageField}
}

func ProjVals(name string, amount int, contactID1, contactID2, flowID string) map[string]interface{} {
	dealVals := map[string]interface{}{
		"uuid-00-projecy-name": name,
		"uuid-00-contacts":     []interface{}{contactID1, contactID2},
		"uuid-00-pipe":         []interface{}{flowID},
		"uuid-00-pipe-stage":   []interface{}{},
		"uuid-00-company":      []interface{}{},
	}
	return dealVals
}

func TaskFields(contactEntityID, companyEntityID, dealEntityID, statusEntityID, nodeEntityID string, stItem1, stItem2, stItem3 string, typeEntityID, typeItemEmailID, typeItemTodoID string, emailEntityID string) []entity.Field {
	descField := entity.Field{
		Key:         "uuid-00-desc",
		Name:        "desc",
		DisplayName: "Notes",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
	}

	dueByField := entity.Field{
		Key:         "uuid-00-due-by",
		Name:        "due_by",
		DisplayName: "Due By",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
	}

	contactField := entity.Field{
		Key:         "uuid-00-contact",
		Name:        "contact",
		DisplayName: "Associated Contact",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       contactEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: "uuid-00-fname"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	projectField := entity.Field{
		Key:         "uuid-00-project",
		Name:        "project",
		DisplayName: "Associated Project",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       dealEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: "uuid-00-project-name"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	companyField := entity.Field{
		Key:         "uuid-00-company",
		Name:        "company",
		DisplayName: "Associated Company",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       companyEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: "uuid-00-name"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	reminderField := entity.Field{
		Key:         "uuid-00-reminder",
		Name:        "reminder",
		DisplayName: "Reminder",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
	}

	statusFieldKey := "uuid-00-status"
	statusField := entity.Field{
		Key:         statusFieldKey,
		Name:        "status",
		DisplayName: "Status",
		DomType:     entity.DomAutoSelect,
		DataType:    entity.TypeReference,
		RefID:       statusEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: "uuid-00-name"},
		Dependent: &entity.Dependent{
			ParentKey:   dueByField.Key,
			Expressions: []string{fmt.Sprintf("{{%s.%s}} in {%s}", "self", statusFieldKey, stItem2), fmt.Sprintf("{{%s.%s}} af {now}", "self", dueByField.Key), fmt.Sprintf("{{%s.%s}} bf {now}", "self", dueByField.Key)},
			Actions: []string{fmt.Sprintf("{{{%s.%s	}}}", reference.ActionSet, reference.ByDoNothing), fmt.Sprintf("{{{%s.%s}}}", reference.ActionSet, stItem1), fmt.Sprintf("{{{%s.%s}}}", reference.ActionSet, stItem3)},
		},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	stageField := entity.Field{
		Key:      "uuid-00-pipe-stage",
		Name:     "pipeline_stage",
		DomType:  entity.DomNotApplicable,
		DataType: entity.TypeReference,
		RefID:    nodeEntityID,
		Meta:     map[string]string{"node": "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	typeField := entity.Field{
		Key:         "uuid-00-type",
		Name:        "type",
		DisplayName: "Type",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       typeEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: "uuid-00-name", entity.MetaKeyLoadChoices: "true"},
		Value:       []interface{}{typeItemTodoID},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	emailTemplateField := entity.Field{
		Key:         "uuid-00-mail-template",
		Name:        "template",
		DisplayName: "Template",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       emailEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: "name"},
		Dependent: &entity.Dependent{
			ParentKey:   typeField.Key,
			Expressions: []string{fmt.Sprintf("{{%s.%s}} in {%s}", "self", typeField.Key, typeItemEmailID), fmt.Sprintf("{{%s.%s}} !in {%s}", "self", typeField.Key, typeItemEmailID)},
			Actions:     []string{fmt.Sprintf("{{{%s.%s}}}", reference.ActionView, reference.ByShow), fmt.Sprintf("{{{%s.%s}}}", reference.ActionView, reference.ByHide)},
		},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{descField, statusField, contactField, projectField, companyField, dueByField, reminderField, stageField, typeField, emailTemplateField}
}

func TaskVals(desc, contactID string, typeID string) map[string]interface{} {
	taskVals := map[string]interface{}{
		"uuid-00-desc":          desc,
		"uuid-00-contact":       []interface{}{contactID},
		"uuid-00-status":        []interface{}{},
		"uuid-00-company":       []interface{}{},
		"uuid-00-deal":          []interface{}{},
		"uuid-00-reminder":      util.FormatTimeGo(time.Now()),
		"uuid-00-due-by":        util.FormatTimeGo(time.Now()),
		"uuid-00-type":          []interface{}{typeID},
		"uuid-00-mail-template": []interface{}{},
		"uuid-00-pipe-stage":    []interface{}{},
	}
	return taskVals
}

func MeetingFields(contactEntityID, companyEntityID, projectEntityID string) []entity.Field {

	titleField := entity.Field{
		Key:         "uuid-00-cal-title",
		Name:        "cal_title",
		DisplayName: "Title",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
	}

	summaryField := entity.Field{
		Key:         "uuid-00-summary",
		Name:        "summary",
		DisplayName: "Summary",
		DomType:     entity.DomTextArea,
		DataType:    entity.TypeString,
	}

	attendessField := entity.Field{
		Key:         "uuid-00-attendess",
		Name:        "attendess",
		DisplayName: "Attendess",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       contactEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: "uuid-00-email"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	startTimeField := entity.Field{
		Key:         "uuid-00-start-time",
		Name:        "start_time",
		DisplayName: "Start Time",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
	}

	endTimeField := entity.Field{
		Key:         "uuid-00-end-time",
		Name:        "end_time",
		DisplayName: "End Time",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
		Meta:        map[string]string{entity.MetaKeyRow: "true"},
	}

	timezoneField := entity.Field{
		Key:         "uuid-00-timezone",
		Name:        "timezone",
		DisplayName: "Timezone",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyHidden: "true"},
	}

	createdAtField := entity.Field{
		Key:         "uuid-00-created-at",
		Name:        "created_at",
		DisplayName: "Created At",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
		Meta:        map[string]string{entity.MetaKeyHidden: "true"},
	}

	updatedAtField := entity.Field{
		Key:         "uuid-00-updated-at",
		Name:        "updated_at",
		DisplayName: "Updated At",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
		Meta:        map[string]string{entity.MetaKeyHidden: "true"},
	}

	contactField := entity.Field{
		Key:         "uuid-00-contact",
		Name:        "contact",
		DisplayName: "Associated Contact",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       contactEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: "uuid-00-fname"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	companyField := entity.Field{
		Key:         "uuid-00-company",
		Name:        "company",
		DisplayName: "Associated Company",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       companyEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: "uuid-00-name", entity.MetaKeyRow: "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	projectField := entity.Field{
		Key:         "uuid-00-project",
		Name:        "project",
		DisplayName: "Associated Project",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       projectEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: "uuid-00-project-name", entity.MetaKeyRow: "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{titleField, summaryField, attendessField, startTimeField, endTimeField, timezoneField, createdAtField, updatedAtField, contactField, companyField, projectField}
}
