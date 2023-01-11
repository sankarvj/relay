package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func TaskFields(contactEntityID, contactEntityKey, companyEntityID, companyEntityKey, nodeEntityID, statusEntityID, statusEntityKey string, ownerEntityID, ownerEntitySearchKey string) []entity.Field {

	nameFieldID := uuid.New().String()
	nameField := entity.Field{
		Key:         nameFieldID,
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
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

	contactFieldID := uuid.New().String()
	contactField := entity.Field{
		Key:         contactFieldID,
		Name:        "associated_contacts",
		DisplayName: "Associated Contacts",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       contactEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: contactEntityKey},
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

	// approverFieldID := uuid.New().String()
	// approverField := entity.Field{
	// 	Key:         approverFieldID,
	// 	Name:        "approver",
	// 	DisplayName: "Approver",
	// 	DomType:     entity.DomSelect,
	// 	DataType:    entity.TypeReference,
	// 	RefID:       ownerEntityID,
	// 	Who:         entity.WhoApprover,
	// 	Meta:        map[string]string{entity.MetaKeyDisplayGex: ownerEntitySearchKey, entity.MetaKeyHidden: "true"},
	// 	Field: &entity.Field{
	// 		DataType: entity.TypeString,
	// 		Key:      "id",
	// 		Value:    "--",
	// 	},
	// }

	return []entity.Field{nameField, descField, statusField, contactField, companyField, dueByField, reminderField, stageField, ownerField}
}
