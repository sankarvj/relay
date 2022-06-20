package em

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func RoleFields() []entity.Field {
	roleNameFieldID := uuid.New().String()
	roleNameField := entity.Field{
		Key:         roleNameFieldID,
		Name:        "role",
		DisplayName: "Role",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	return []entity.Field{roleNameField}
}

func RoleVals(namekey, name string) map[string]interface{} {
	statusVals := map[string]interface{}{
		namekey: name,
	}
	return statusVals
}

func EmployeeFields(ownerEntityID, ownerEntityKey string, roleEntityID, roleEntityKey string) []entity.Field {
	nameFieldID := uuid.New().String()
	nameField := entity.Field{
		Key:         nameFieldID,
		Name:        "first_name",
		DisplayName: "First Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle, entity.MetaKeyUnique: "true"},
	}

	personalEmailFieldID := uuid.New().String()
	personalEmailField := entity.Field{
		Key:         personalEmailFieldID,
		Name:        "email",
		DisplayName: "Personal Email",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Who:         entity.WhoEmail,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle},
	}

	mobileFieldID := uuid.New().String()
	mobileField := entity.Field{
		Key:         mobileFieldID,
		Name:        "mobile_numbers",
		DisplayName: "Mobile Numbers",
		DataType:    entity.TypeList,
		DomType:     entity.DomMultiSelect,
		Field: &entity.Field{
			Key:      "element",
			DataType: entity.TypeString,
		},
	}

	avatarFieldID := uuid.New().String()
	avatarField := entity.Field{
		Key:         avatarFieldID,
		Name:        "avatar",
		DisplayName: "Avatar",
		DataType:    entity.TypeString,
		DomType:     entity.DomImage,
		Who:         entity.WhoAvatar,
	}

	officeEmailFieldID := uuid.New().String()
	officeEmailField := entity.Field{
		Key:         officeEmailFieldID,
		Name:        "email",
		DisplayName: "Work Email",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Who:         entity.WhoEmail,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle},
	}

	lfStageFieldID := uuid.New().String()
	lfStageField := entity.Field{
		Key:         lfStageFieldID,
		Name:        "state",
		DisplayName: "State",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeList,
		Choices: []entity.Choice{
			{
				ID:           "0",
				DisplayValue: "Onboarding",
			},
			{
				ID:           "1",
				DisplayValue: "Active",
			},
			{
				ID:           "2",
				DisplayValue: "Quit",
			},
		},
		Field: &entity.Field{
			Key:      "element",
			DataType: entity.TypeString,
		},
	}

	ownerFieldID := uuid.New().String()
	ownerField := entity.Field{
		Key:         ownerFieldID,
		Name:        "hr",
		DisplayName: "HR",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       ownerEntityID,
		RefType:     entity.RefTypeStraight,
		Who:         entity.WhoAssignee,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: ownerEntityKey, entity.MetaKeyLayout: entity.MetaLayoutUsers},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	managerFieldID := uuid.New().String()
	managerField := entity.Field{
		Key:         managerFieldID,
		Name:        "manager",
		DisplayName: "Manager",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       ownerEntityID,
		RefType:     entity.RefTypeStraight,
		Who:         entity.WhoAssignee,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: ownerEntityKey, entity.MetaKeyLayout: entity.MetaLayoutUsers},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	roleFieldID := uuid.New().String()
	roleField := entity.Field{
		Key:         roleFieldID,
		Name:        "role",
		DisplayName: "Role",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       roleEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: roleEntityKey},
		Who:         entity.WhoStatus,
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{nameField, personalEmailField, mobileField, officeEmailField, lfStageField, avatarField, ownerField, managerField, roleField}
}

func AssetStatusFields() []entity.Field {
	nameFieldID := uuid.New().String()
	nameField := entity.Field{
		Key:         nameFieldID,
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	colorFieldID := uuid.New().String()
	colorField := entity.Field{
		Key:         colorFieldID,
		Name:        "color",
		DisplayName: "Color",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	return []entity.Field{nameField, colorField}
}

func AssetStatusVals(namekey, colorKey, name, color string) map[string]interface{} {
	statusVals := map[string]interface{}{
		namekey:  name,
		colorKey: color,
	}
	return statusVals
}

func AssetRequestFields(assetEntityID string, assetEntityKey string, assetStatusEntityID, assestStatusEntityKey string) []entity.Field {
	assetFieldID := uuid.New().String()
	assetField := entity.Field{
		Key:         assetFieldID,
		Name:        "asset",
		DisplayName: "Asset",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       assetEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: assetEntityKey, entity.MetaMultiChoice: "false"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	commentsFieldID := uuid.New().String()
	commentsField := entity.Field{
		Key:         commentsFieldID,
		Name:        "comments",
		DisplayName: "Comments",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle, entity.MetaKeyHTML: "true"},
	}

	statusFieldID := uuid.New().String()
	statusField := entity.Field{
		Key:         statusFieldID,
		Name:        "request_status",
		DisplayName: "Request Status",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       assetStatusEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: assestStatusEntityKey},
		Who:         entity.WhoStatus,
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{assetField, commentsField, statusField}
}

func ServiceRequestFields(serviceEntityID string, serviceEntityKey string, statusEntityID, statusEntityKey string) []entity.Field {
	assetFieldID := uuid.New().String()
	assetField := entity.Field{
		Key:         assetFieldID,
		Name:        "service",
		DisplayName: "Service",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       serviceEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: serviceEntityKey, entity.MetaMultiChoice: "false"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
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

	statusFieldID := uuid.New().String()
	statusField := entity.Field{
		Key:         statusFieldID,
		Name:        "request_status",
		DisplayName: "Request Status",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       statusEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: statusEntityKey},
		Who:         entity.WhoStatus,
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{assetField, descField, statusField}
}
