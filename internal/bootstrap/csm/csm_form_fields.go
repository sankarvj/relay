package csm

import (
	"fmt"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/reference"
)

func ProjectFields(contactEntityID, companyEntityID string, flowEntityID, nodeEntityID string) []entity.Field {
	projectName := entity.Field{
		Key:         "uuid-00-project-name",
		Name:        "project_name",
		DisplayName: "Project Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "title"},
	}

	contactsField := entity.Field{
		Key:         "uuid-00-contacts",
		Name:        "contact",
		DisplayName: "Associated Contacts",
		DomType:     entity.DomMultiSelect,
		DataType:    entity.TypeReference,
		RefID:       contactEntityID,
		Meta:        map[string]string{"display_gex": "uuid-00-fname"},
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
		Meta:        map[string]string{"display_gex": "uuid-00-name"},
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
		Meta:        map[string]string{"flow": "true"},
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
		Meta: map[string]string{"display_gex": "uuid-00-fname", "node": "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      pipeField.Key,
			Value:    "--",
		},
	}

	return []entity.Field{projectName, contactsField, companyField, pipeField, pipeStageField}
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
		Meta:        map[string]string{"layout": "title"},
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
		Meta:        map[string]string{"display_gex": "uuid-00-fname"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	dealField := entity.Field{
		Key:         "uuid-00-deal",
		Name:        "deal",
		DisplayName: "Associated Deal",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       dealEntityID,
		Meta:        map[string]string{"display_gex": "uuid-00-deal-name"},
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
		Meta:        map[string]string{"display_gex": "uuid-00-name"},
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
		Meta:        map[string]string{"display_gex": "uuid-00-name"},
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
		Meta:        map[string]string{"display_gex": "uuid-00-name", "load_choices": "true"},
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
		Meta:        map[string]string{"display_gex": "name"},
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

	return []entity.Field{descField, statusField, contactField, dealField, companyField, dueByField, reminderField, stageField, typeField, emailTemplateField}
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
		DisplayName: "Verb",
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
		DisplayName: "Verb",
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
