package base

import (
	"fmt"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/reference"
)

func FlowFields() []entity.Field {
	actualFlowID := entity.Field{
		Key:         "uuid-00-flow-id",
		Name:        "flow_id",
		DisplayName: "ID",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
	}

	return []entity.Field{actualFlowID}
}

func NodeFields() []entity.Field {
	actualFlowID := entity.Field{
		Key:         "uuid-00-node-id",
		Name:        "node_id",
		DisplayName: "ID",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
	}

	return []entity.Field{actualFlowID}
}

func StatusFields() []entity.Field {
	verbField := entity.Field{
		Key:         entity.VerbKey, // we use this value inside the code. don't change it
		Name:        entity.Verb,
		DisplayName: "Verb (Internal field)",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
	}

	nameField := entity.Field{
		Key:         "uuid-00-name",
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	colorField := entity.Field{
		Key:         "uuid-00-color",
		Name:        "color",
		DisplayName: "Color",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	return []entity.Field{verbField, nameField, colorField}
}

func StatusVals(verb, name, color string) map[string]interface{} {
	statusVals := map[string]interface{}{
		entity.VerbKey:  verb,
		"uuid-00-name":  name,
		"uuid-00-color": color,
	}
	return statusVals
}

func TypeFields() []entity.Field {
	verbField := entity.Field{
		Key:         entity.VerbKey, // we use this value inside the code. don't change it
		Name:        entity.Verb,
		DisplayName: "Verb (Internal field)",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
	}

	nameField := entity.Field{
		Key:         "uuid-00-name",
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	return []entity.Field{verbField, nameField}
}

func TypeVals(verb, name string) map[string]interface{} {
	typeVals := map[string]interface{}{
		entity.VerbKey: verb,
		"uuid-00-name": name,
	}
	return typeVals
}

func APIFields() []entity.Field {
	path := entity.Field{
		Key:      "uuid-00-path",
		Name:     "path",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "/actuator/info",
	}

	host := entity.Field{
		Key:      "uuid-00-host",
		Name:     "host",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "https://stage.freshcontacts.io",
	}

	method := entity.Field{
		Key:      "uuid-00-method",
		Name:     "method",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "GET",
	}

	headers := entity.Field{
		Key:      "uuid-00-headers",
		Name:     "headers",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "{\"X-ClientToken\":\"mcr eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjExNTc2NTAwMjk2fQ.1KtXw_YgxbJW8ibv_v2hfpInjQKC6enCh9IO1ziV2RA\"}",
	}

	return []entity.Field{path, host, method, headers}
}

func DelayVals() map[string]interface{} {
	delayVals := map[string]interface{}{
		"uuid-00-title":    "1 hr delay",
		"uuid-00-delay-by": 1,
		"uuid-00-repeat":   "false",
	}
	return delayVals
}

func DelayFields() []entity.Field {

	titleField := entity.Field{
		Key:         "uuid-00-title",
		Name:        "title",
		DisplayName: "Title",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
	}

	delay_by := entity.Field{
		Key:         "uuid-00-delay-by",
		Name:        "delay_by",
		DisplayName: "Delay By",
		DomType:     entity.DomText,
		DataType:    entity.TypeNumber,
	}

	repeat := entity.Field{
		Key:         "uuid-00-repeat",
		Name:        "repeat",
		DisplayName: "Repeat",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Value:       "true",
	}

	return []entity.Field{titleField, delay_by, repeat}
}

func MeetingFields(contactEntityID, companyEntityID string) []entity.Field {

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

	return []entity.Field{titleField, summaryField, attendessField, startTimeField, endTimeField, timezoneField, createdAtField, updatedAtField, contactField, companyField}
}

func NoteFields(contactEntityID, companyEntityID string) []entity.Field {
	descField := entity.Field{
		Key:         "uuid-00-desc",
		Name:        "desc",
		DisplayName: "Notes",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	contactField := entity.Field{
		Key:         "uuid-00-contact",
		Name:        "contact",
		DisplayName: "Associated To",
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
		DisplayName: "Associated To",
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

	return []entity.Field{descField, contactField, companyField}
}

func TaskFields(contactEntityID, companyEntityID, statusEntityID, nodeEntityID string, stItem1, stItem2, stItem3 string, typeEntityID, typeItemEmailID, typeItemTodoID string, emailEntityID string) []entity.Field {
	descField := entity.Field{
		Key:         "uuid-00-desc",
		Name:        "desc",
		DisplayName: "Notes",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title", entity.MetaKeyHTML: "true"},
	}

	dueByField := entity.Field{
		Key:         "uuid-00-due-by",
		Name:        "due_by",
		DisplayName: "Due By",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
		Who:         entity.WhoDueBy,
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
		Who:         entity.WhoReminder,
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
		Who:         entity.WhoStatus,
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
		Meta:     map[string]string{entity.MetaKeyNode: "true"},
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

	return []entity.Field{descField, statusField, contactField, companyField, dueByField, reminderField, stageField, typeField, emailTemplateField}
}

func TicketFields(contactEntityID, companyEntityID, statusEntityID string) []entity.Field {
	nameField := entity.Field{
		Key:         "uuid-00-subject",
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
	}

	statusField := entity.Field{
		Key:         "uuid-00-status",
		Name:        "status",
		DisplayName: "Status",
		DomType:     entity.DomSelect,
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

	contactField := entity.Field{
		Key:         "uuid-00-contact",
		Name:        "contact",
		DisplayName: "Associated Contacts",
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

	return []entity.Field{nameField, statusField, contactField, companyField}
}

func TicketVals(name, statusID string) map[string]interface{} {
	ticketVals := map[string]interface{}{
		"uuid-00-subject": name,
		"uuid-00-contact": []interface{}{},
		"uuid-00-company": []interface{}{},
		"uuid-00-status":  []interface{}{statusID},
	}
	return ticketVals
}
