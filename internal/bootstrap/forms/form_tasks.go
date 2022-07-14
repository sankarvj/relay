package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func TaskFields(contactEntityID, companyEntityID, nodeEntityID, statusEntityID string, ownerEntityID, ownerEntitySearchKey string) []entity.Field {

	descFieldID := uuid.New().String()
	descField := entity.Field{
		Key:         descFieldID,
		Name:        "desc",
		DisplayName: "Notes",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title", entity.MetaKeyHTML: "true"},
	}

	contactFieldID := uuid.New().String()
	contactField := entity.Field{
		Key:         contactFieldID,
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

	companyFieldID := uuid.New().String()
	companyField := entity.Field{
		Key:         companyFieldID,
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

	dueByFieldID := uuid.New().String()
	dueByField := entity.Field{
		Key:         dueByFieldID,
		Name:        "due_by",
		DisplayName: "Due By",
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
		DomType:     entity.DomAutoSelect,
		DataType:    entity.TypeReference,
		RefID:       statusEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: "uuid-00-name"},
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

	stageFieldID := uuid.New().String()
	stageField := entity.Field{
		Key:      stageFieldID,
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

	ownersFieldID := uuid.New().String()
	ownerField := entity.Field{
		Key:         ownersFieldID,
		Name:        "assignees",
		DisplayName: "Assignees",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       ownerEntityID,
		Who:         entity.WhoAssignee,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: ownerEntitySearchKey, entity.MetaKeyLayout: entity.MetaLayoutUsers},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{descField, statusField, contactField, companyField, dueByField, reminderField, stageField, ownerField}
}
