package incident

import (
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/reference"
)

func IncidentFields(statusEntityID, statusEntityKey, priorityEntityID, priorityTitleKey, typeEntityID, typeTitleKey, catEntityID, catTitleKey, ownerEntityID, ownerEntityKey string, flowEntityID, nodeEntityID, nodeKey string) []entity.Field {
	incidentNameFieldID := uuid.New().String()
	incidentNameField := entity.Field{
		Key:         incidentNameFieldID,
		Name:        "incident_name",
		DisplayName: "Incident Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	incidentDescFieldID := uuid.New().String()
	incidentDescField := entity.Field{
		Key:         incidentDescFieldID,
		Name:        "incident_desc",
		DisplayName: "Description",
		DomType:     entity.DomTextArea,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle},
	}

	tagFieldID := uuid.New().String()
	tagField := entity.Field{
		Key:         tagFieldID,
		Name:        "tags",
		DisplayName: "Tags",
		DataType:    entity.TypeList,
		DomType:     entity.DomMultiSelect,
		Field: &entity.Field{
			Key:      "id",
			DataType: entity.TypeString,
		},
	}

	priorityFieldID := uuid.New().String()
	priorityField := entity.Field{
		Key:         priorityFieldID,
		Name:        "priority",
		DisplayName: "Priority",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       priorityEntityID,
		RefType:     entity.RefTypeStraight,
		Who:         entity.WhoPriority,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: priorityTitleKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	typeFieldID := uuid.New().String()
	typeField := entity.Field{
		Key:         typeFieldID,
		Name:        "type",
		DisplayName: "Type",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       typeEntityID,
		RefType:     entity.RefTypeStraight,
		Who:         entity.WhoType,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: typeTitleKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	categoryFieldID := uuid.New().String()
	categoryField := entity.Field{
		Key:         categoryFieldID,
		Name:        "category",
		DisplayName: "Category",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       catEntityID,
		RefType:     entity.RefTypeStraight,
		Who:         entity.WhoCategory,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: catTitleKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	followersFieldID := uuid.New().String()
	followerField := entity.Field{
		Key:         followersFieldID,
		Name:        "followers",
		DisplayName: "Followers",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       ownerEntityID,
		Who:         entity.WhoFollower,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: ownerEntityKey, entity.MetaMultiChoice: "true", entity.MetaKeyLayout: entity.MetaLayoutUsers},
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
		DisplayName: "Incident Owner",
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

	return []entity.Field{incidentNameField, incidentDescField, tagField, priorityField, typeField, categoryField, ownerField, followerField, startTimeField, endTimeField, pipeField, pipeStageField}
}

func AlertFields(statusEntityID, statusEntityKey, priorityEntityID, priorityTitleKey, typeEntityID, typeTitleKey, catEntityID, catTitleKey, ownerEntityID, ownerEntityKey string, flowEntityID, nodeEntityID, nodeKey string) []entity.Field {
	alertNameFieldID := uuid.New().String()
	alertNameField := entity.Field{
		Key:         alertNameFieldID,
		Name:        "alert_name",
		DisplayName: "Alert Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	occurrenceFieldID := uuid.New().String()
	occurrenceField := entity.Field{
		Key:         occurrenceFieldID,
		Name:        "occurrence",
		DisplayName: "occurrence",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeNumber,
		Meta:        map[string]string{entity.MetaKeyHidden: "true"},
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

	priorityFieldID := uuid.New().String()
	priorityField := entity.Field{
		Key:         priorityFieldID,
		Name:        "priority",
		DisplayName: "Priority",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       priorityEntityID,
		RefType:     entity.RefTypeStraight,
		Who:         entity.WhoPriority,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: priorityTitleKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	typeFieldID := uuid.New().String()
	typeField := entity.Field{
		Key:         typeFieldID,
		Name:        "type",
		DisplayName: "Type",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       typeEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: typeTitleKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	categoryFieldID := uuid.New().String()
	categoryField := entity.Field{
		Key:         categoryFieldID,
		Name:        "category",
		DisplayName: "Category",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       catEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: catTitleKey},
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
		DisplayName: "Incident Owner",
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

	return []entity.Field{alertNameField, statusField, priorityField, typeField, occurrenceField, categoryField, ownerField, pipeField, pipeStageField}
}

func BugFields(statusEntityID, statusEntityKey, priorityEntityID, priorityTitleKey, typeEntityID, typeTitleKey, catEntityID, catTitleKey, ownerEntityID, ownerEntityKey string, flowEntityID, nodeEntityID, nodeKey string) []entity.Field {
	bugNameFieldID := uuid.New().String()
	bugNameField := entity.Field{
		Key:         bugNameFieldID,
		Name:        "bug_name",
		DisplayName: "Bug Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	bugDescFieldID := uuid.New().String()
	bugDescField := entity.Field{
		Key:         bugDescFieldID,
		Name:        "bug_desc",
		DisplayName: "Description",
		DomType:     entity.DomTextArea,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle},
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

	priorityFieldID := uuid.New().String()
	priorityField := entity.Field{
		Key:         priorityFieldID,
		Name:        "priority",
		DisplayName: "Priority",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       priorityEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: priorityTitleKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	typeFieldID := uuid.New().String()
	typeField := entity.Field{
		Key:         typeFieldID,
		Name:        "type",
		DisplayName: "Type",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       typeEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: typeTitleKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	categoryFieldID := uuid.New().String()
	categoryField := entity.Field{
		Key:         categoryFieldID,
		Name:        "category",
		DisplayName: "Category",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       catEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: catTitleKey},
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
		DisplayName: "Incident Owner",
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

	return []entity.Field{bugNameField, bugDescField, statusField, priorityField, typeField, categoryField, ownerField, pipeField, pipeStageField}
}

func PriorityFields() []entity.Field {
	verbFieldID := uuid.New().String()
	verbField := entity.Field{
		Key:         verbFieldID,
		Name:        "verb",
		DisplayName: "Verb (Internal field)",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
	}

	nameFieldID := uuid.New().String()
	nameField := entity.Field{
		Key:         nameFieldID,
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
	}

	colorFieldID := uuid.New().String()
	colorField := entity.Field{
		Key:         colorFieldID,
		Name:        "color",
		DisplayName: "Color",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Who:         entity.WhoColor,
	}

	return []entity.Field{verbField, nameField, colorField}
}

func PriorityVals(nameKey, colorKey, name, color string) map[string]interface{} {
	priorityVals := map[string]interface{}{
		nameKey:  name,
		colorKey: color,
	}
	return priorityVals
}

func TypeFields() []entity.Field {
	verbFieldID := uuid.New().String()
	verbField := entity.Field{
		Key:         verbFieldID,
		Name:        "verb",
		DisplayName: "Verb (Internal field)",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
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

	nameFieldID := uuid.New().String()
	nameField := entity.Field{
		Key:         nameFieldID,
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
	}

	colorFieldID := uuid.New().String()
	colorField := entity.Field{
		Key:         colorFieldID,
		Name:        "color",
		DisplayName: "Color",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Who:         entity.WhoColor,
	}

	return []entity.Field{verbField, iconField, nameField, colorField}
}

func TypeVals(iconKey, nameKey, colorKey, icon, name, color string) map[string]interface{} {
	typeVals := map[string]interface{}{
		iconKey:  icon,
		nameKey:  name,
		colorKey: color,
	}
	return typeVals
}

func TaskFields(statusEntityID, statusEntityKey, ownerEntityID, ownerEntityKey string) []entity.Field {
	nameFieldID := uuid.New().String()
	nameField := entity.Field{
		Key:         nameFieldID,
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
	}

	descFieldID := uuid.New().String()
	descField := entity.Field{
		Key:         descFieldID,
		Name:        "desc",
		DisplayName: "Description",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle, entity.MetaKeyHTML: "true"},
	}

	dueByFieldID := uuid.New().String()
	dueByField := entity.Field{
		Key:         dueByFieldID,
		Name:        "due_by",
		DisplayName: "Due by",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
		Who:         entity.WhoDueBy,
	}

	reminderFieldID := uuid.New().String()
	reminderField := entity.Field{
		Key:         reminderFieldID,
		Name:        "reminder",
		DisplayName: "Reminder",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
		Who:         entity.WhoReminder,
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
		Who:         entity.WhoStatus,
		Dependent: &entity.Dependent{
			ParentKey: dueByField.Key,
		},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	ownerFieldID := uuid.New().String()
	ownerField := entity.Field{
		Key:         ownerFieldID,
		Name:        "assignee",
		DisplayName: "Assignee",
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

	return []entity.Field{nameField, descField, dueByField, reminderField, statusField, ownerField}
}

func NoteFields() []entity.Field {
	descFieldID := uuid.New().String()
	descField := entity.Field{
		Key:         descFieldID,
		Name:        "desc",
		DisplayName: "Notes",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	return []entity.Field{descField}
}

func NoteVals(noteEntity entity.Entity, desc string) map[string]interface{} {
	noteVals := map[string]interface{}{
		"desc": desc,
	}
	return forms.KeyMap(noteEntity.NameKeyMapWrapper(), noteVals)
}
