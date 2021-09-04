package crm

import (
	"fmt"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"gitlab.com/vjsideprojects/relay/internal/schema"
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

func ContactFields(statusEntityID, ownerEntityID string, ownerEntityKey string) []entity.Field {
	nameField := entity.Field{
		Key:         schema.SeedFieldFNameKey,
		Name:        "first_name",
		DisplayName: "First Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
	}

	emailField := entity.Field{
		Key:         "uuid-00-email",
		Name:        "email",
		DisplayName: "Email",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "sub-title"},
	}

	mobileField := entity.Field{
		Key:         "uuid-00-mobile-numbers",
		Name:        "mobile_numbers",
		DisplayName: "Mobile Numbers",
		DataType:    entity.TypeList,
		DomType:     entity.DomMultiSelect,
		Field: &entity.Field{
			Key:      "element",
			DataType: entity.TypeString,
		},
	}

	npsField := entity.Field{
		Key:         schema.SeedFieldNPSKey,
		Name:        "nps_score",
		DisplayName: "NPS Score",
		DataType:    entity.TypeNumber,
		DomType:     entity.DomText,
	}

	lfStageField := entity.Field{
		Key:         "uuid-00-lf-stage",
		Name:        "lifecycle_stage",
		DisplayName: "Lifecycle Stage",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeList,
		Choices: []entity.Choice{
			{
				ID:           "1",
				DisplayValue: "Lead",
			},
			{
				ID:           "2",
				DisplayValue: "Contact",
			},
		},
		Field: &entity.Field{
			Key:      "element",
			DataType: entity.TypeString,
		},
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

	ownerField := entity.Field{
		Key:         "uuid-00-owner",
		Name:        "owner",
		DisplayName: "Owner",
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

	return []entity.Field{nameField, emailField, mobileField, npsField, lfStageField, statusField, ownerField}
}

func ContactVals(name, email, statusID string) map[string]interface{} {
	contactVals := map[string]interface{}{
		schema.SeedFieldFNameKey: name,
		"uuid-00-email":          email,
		"uuid-00-mobile-numbers": []interface{}{"9944293499", "9940209164"},
		schema.SeedFieldNPSKey:   100,
		"uuid-00-lf-stage":       []interface{}{"1"},
		"uuid-00-status":         []interface{}{statusID},
		"uuid-00-owner":          []interface{}{},
	}
	return contactVals
}

func CompanyFields(ownerEntityID string, ownerEntityKey string) []entity.Field {
	nameField := entity.Field{
		Key:         "uuid-00-name",
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
	}

	websiteField := entity.Field{
		Key:         "uuid-00-website",
		Name:        "website",
		DisplayName: "Domain",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	cityField := entity.Field{
		Key:         "uuid-00-city",
		Name:        "city",
		DisplayName: "City",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	stateField := entity.Field{
		Key:         "uuid-00-state",
		Name:        "state",
		DisplayName: "State",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	annualRevenueField := entity.Field{
		Key:         "uuid-00-revenue",
		Name:        "revenue",
		DisplayName: "Annual Revenue",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	countryField := entity.Field{
		Key:         "uuid-00-country",
		Name:        "country",
		DisplayName: "Country",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	employeesCountField := entity.Field{
		Key:         "uuid-00-employees-count",
		Name:        "employees_count",
		DisplayName: "Employees Count",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	ownerField := entity.Field{
		Key:         "uuid-00-owner",
		Name:        "owner",
		DisplayName: "Company Owner",
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

	return []entity.Field{nameField, websiteField, cityField, stateField, ownerField, annualRevenueField, countryField, employeesCountField}
}

func CompanyVals(name, website string) map[string]interface{} {
	companyVals := map[string]interface{}{
		"uuid-00-name":    name,
		"uuid-00-website": website,
		"uuid-00-owner":   []interface{}{},
	}
	return companyVals
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

	dealField := entity.Field{
		Key:         "uuid-00-deal",
		Name:        "deal",
		DisplayName: "Associated Deal",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       dealEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: "uuid-00-deal-name"},
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

func MeetingFields(contactEntityID, companyEntityID, dealEntityID string) []entity.Field {

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

	dealField := entity.Field{
		Key:         "uuid-00-deal",
		Name:        "deal",
		DisplayName: "Associated Deal",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       dealEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: "uuid-00-deal-name", entity.MetaKeyRow: "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{titleField, summaryField, attendessField, startTimeField, endTimeField, timezoneField, createdAtField, updatedAtField, contactField, companyField, dealField}
}

func DealFields(contactEntityID, companyEntityID string, flowEntityID, nodeEntityID string) []entity.Field {
	dealName := entity.Field{
		Key:         "uuid-00-deal-name",
		Name:        "deal_name",
		DisplayName: "Deal Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
	}

	dealAmount := entity.Field{
		Key:         "uuid-00-deal-amount",
		Name:        "deal_amount",
		DisplayName: "Deal Amount",
		DomType:     entity.DomText,
		DataType:    entity.TypeNumber,
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
		Meta: map[string]string{entity.MetaKeyDisplayGex: "uuid-00-fname", entity.MetaKeyNode: "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      pipeField.Key,
			Value:    "--",
		},
	}

	return []entity.Field{dealName, dealAmount, contactsField, companyField, pipeField, pipeStageField}
}

func DealVals(name string, amount int, contactID1, contactID2, flowID string) map[string]interface{} {
	dealVals := map[string]interface{}{
		"uuid-00-deal-name":   name,
		"uuid-00-deal-amount": amount,
		"uuid-00-contacts":    []interface{}{contactID1, contactID2},
		"uuid-00-pipe":        []interface{}{flowID},
		"uuid-00-pipe-stage":  []interface{}{},
		"uuid-00-company":     []interface{}{},
	}
	return dealVals
}

func NoteFields(contactEntityID, companyEntityID, dealEntityID string) []entity.Field {
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

	dealField := entity.Field{
		Key:         "uuid-00-deal",
		Name:        "deal",
		DisplayName: "Associated To",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       dealEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: "uuid-00-deal-name"},
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

	return []entity.Field{descField, contactField, dealField, companyField}
}

func NoteVals(desc, contactID string) map[string]interface{} {
	noteVals := map[string]interface{}{
		"uuid-00-desc":    desc,
		"uuid-00-contact": []interface{}{contactID},
	}
	return noteVals
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
		"uuid-00-delay-by": "2",
		"uuid-00-repeat":   "true",
	}
	return delayVals
}

func DelayFields() []entity.Field {
	delay_by := entity.Field{
		Key:      "uuid-00-delay-by",
		Name:     "delay_by",
		DomType:  entity.DomText,
		DataType: entity.TypeDateTime,
	}

	repeat := entity.Field{
		Key:      "uuid-00-repeat",
		Name:     "repeat",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "true",
	}

	return []entity.Field{delay_by, repeat}
}
