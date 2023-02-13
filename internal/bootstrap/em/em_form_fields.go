package em

import (
	"fmt"

	"github.com/Pallinder/go-randomdata"
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/reference"
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

func EmployeeFields(flowEntityID, nodeEntityID, nodeKey, ownerEntityID, ownerEntityKey string, roleEntityID, roleEntityKey string) []entity.Field {
	firstNameFieldID := uuid.New().String()
	firstNameField := entity.Field{
		Key:         firstNameFieldID,
		Name:        "first_name",
		DisplayName: "First Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	lastNameFieldID := uuid.New().String()
	lastNameField := entity.Field{
		Key:         lastNameFieldID,
		Name:        "last_name",
		DisplayName: "Last Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	personalEmailFieldID := uuid.New().String()
	personalEmailField := entity.Field{
		Key:         personalEmailFieldID,
		Name:        "personal_email",
		DisplayName: "Personal Email",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Who:         entity.WhoEmail,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle, entity.MetaKeyUnique: "true"},
	}

	mobileFieldID := uuid.New().String()
	mobileField := entity.Field{
		Key:         mobileFieldID,
		Name:        "mobile_numbers",
		DisplayName: "Mobile Numbers",
		DataType:    entity.TypeList,
		DomType:     entity.DomMultiSelect,
		Field: &entity.Field{
			Key:      "id",
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
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle, entity.MetaKeyUnique: "true"},
	}

	lifecyleFieldID := uuid.New().String()
	lifecyleField := entity.Field{
		Key:         lifecyleFieldID,
		Name:        "lifecyle",
		DisplayName: "Lifecycle",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       flowEntityID,
		RefType:     entity.RefTypeStraight,
		Meta:        map[string]string{entity.MetaKeyFlow: "true", entity.MetaMultiChoice: "false", entity.MetaKeyRequired: "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	lifecyleStageFieldID := uuid.New().String()
	lifecyleStageField := entity.Field{
		Key:         lifecyleStageFieldID,
		Name:        "lifecycle_stage",
		DisplayName: "Lifecycle Stage",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       nodeEntityID,
		RefType:     entity.RefTypeStraight,
		Dependent: &entity.Dependent{
			ParentKey:   lifecyleField.Key,
			Expressions: []string{""}, // empty means positive
			Actions:     []string{fmt.Sprintf("{{{%s.%s}}}", reference.ActionFilter, reference.ByFlow)},
		},
		Meta: map[string]string{entity.MetaKeyDisplayGex: nodeKey, entity.MetaKeyNode: "true", entity.MetaKeyRequired: "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      lifecyleField.Key,
			Value:    "--",
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
		DisplayName: "Reporting to",
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
		Meta:        map[string]string{entity.MetaKeyDisplayGex: roleEntityKey, entity.MetaKeyRequired: "true"},
		Who:         entity.WhoStatus,
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	joiningDateFieldID := uuid.New().String()
	joiningDateField := entity.Field{
		Key:         joiningDateFieldID,
		Name:        "joining_date",
		DisplayName: "Joining Date",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
	}

	exitDateFieldID := uuid.New().String()
	exitDateField := entity.Field{
		Key:         exitDateFieldID,
		Name:        "exit_date",
		DisplayName: "Exit date",
		DomType:     entity.DomText,
		DataType:    entity.TypeDate,
	}

	return []entity.Field{firstNameField, lastNameField, personalEmailField, mobileField, officeEmailField, lifecyleField, lifecyleStageField, avatarField, ownerField, managerField, roleField, joiningDateField, exitDateField}
}

func EmpVals(empEntity entity.Entity, firstName, lastName, email, roleID string) map[string]interface{} {
	empVals := map[string]interface{}{
		"first_name":     firstName,
		"last_name":      lastName,
		"email":          email,
		"mobile_numbers": []interface{}{randomdata.PhoneNumber(), randomdata.PhoneNumber()},
		"hr":             []interface{}{},
		"manager":        []interface{}{},
		"avatar":         fmt.Sprintf("https://avatars.dicebear.com/api/pixel-art/%s.svg", firstName),
		"role":           []interface{}{roleID},
	}

	return forms.KeyMap(empEntity.NameKeyMapWrapper(), empVals)
}

func PayrollFields() []entity.Field {
	bankNameFieldID := uuid.New().String()
	bankNameField := entity.Field{
		Key:         bankNameFieldID,
		Name:        "bank_name",
		DisplayName: "Bank Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	accNoFieldID := uuid.New().String()
	accNoField := entity.Field{
		Key:         accNoFieldID,
		Name:        "account",
		DisplayName: "Account Number",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle},
	}

	accountTypeFieldID := uuid.New().String()
	accountTypeField := entity.Field{
		Key:         accountTypeFieldID,
		Name:        "account_type",
		DisplayName: "Account Type",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeList,
		Choices: []entity.Choice{
			{
				ID:           "1",
				DisplayValue: "Savings",
			},
			{
				ID:           "2",
				DisplayValue: "Current",
			},
		},
		Field: &entity.Field{
			Key:      "id",
			DataType: entity.TypeString,
		},
	}

	return []entity.Field{bankNameField, accNoField, accountTypeField}
}

func SalaryFields() []entity.Field {
	baseSalaryFieldID := uuid.New().String()
	baseSalaryField := entity.Field{
		Key:         baseSalaryFieldID,
		Name:        "base_salary",
		DisplayName: "Base Salary",
		DomType:     entity.DomText,
		DataType:    entity.TypeNumber,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	grossSalaryID := uuid.New().String()
	grossSalaryField := entity.Field{
		Key:         grossSalaryID,
		Name:        "gross_salary",
		DisplayName: "Gross Salary",
		DomType:     entity.DomText,
		DataType:    entity.TypeNumber,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle},
	}

	netTaxFieldID := uuid.New().String()
	netTaxField := entity.Field{
		Key:         netTaxFieldID,
		Name:        "net_tax",
		DisplayName: "Net Tax",
		DomType:     entity.DomText,
		DataType:    entity.TypeNumber,
	}

	onFieldID := uuid.New().String()
	onField := entity.Field{
		Key:         onFieldID,
		Name:        "month",
		DisplayName: "Month",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
	}

	return []entity.Field{baseSalaryField, grossSalaryField, netTaxField, onField}
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

func AssetRequestFields(assetEntityID string, assetEntityKey string, assetStatusEntityID, assestStatusEntityKey string, ownerEntityID, ownerEntitySearchKey string) []entity.Field {
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

	dueByFieldID := uuid.New().String()
	dueByField := entity.Field{
		Key:         dueByFieldID,
		Name:        "due_by",
		DisplayName: "Due by",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
		Who:         entity.WhoDueBy,
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

	return []entity.Field{assetField, nameField, descField, statusField, dueByField, ownerField}
}

func ServiceRequestFields(serviceEntityID string, serviceEntityKey string, statusEntityID, statusEntityKey string, ownerEntityID, ownerEntitySearchKey string) []entity.Field {
	serviceFieldID := uuid.New().String()
	serviceField := entity.Field{
		Key:         serviceFieldID,
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

	dueByFieldID := uuid.New().String()
	dueByField := entity.Field{
		Key:         dueByFieldID,
		Name:        "due_by",
		DisplayName: "Due by",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
		Who:         entity.WhoDueBy,
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

	return []entity.Field{serviceField, nameField, descField, statusField, dueByField, ownerField}
}

func TaskEFields(employeeEntityID, employeeEntityKey, nodeEntityID, statusEntityID, statusEntityKey string, ownerEntityID, ownerEntitySearchKey string) []entity.Field {

	descFieldID := uuid.New().String()
	descField := entity.Field{
		Key:         descFieldID,
		Name:        "desc",
		DisplayName: "Notes",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title", entity.MetaKeyHTML: "true"},
	}

	employeeFieldID := uuid.New().String()
	employeeField := entity.Field{
		Key:         employeeFieldID,
		Name:        "employee",
		DisplayName: "Employee",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       employeeEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: employeeEntityKey},
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

	return []entity.Field{descField, statusField, employeeField, dueByField, reminderField, stageField, ownerField}
}
