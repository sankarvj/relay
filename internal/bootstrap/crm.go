package bootstrap

import (
	"fmt"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
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
		Key:         "uuid-00-verb",
		Name:        "verb",
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
		"uuid-00-verb":  verb,
		"uuid-00-name":  name,
		"uuid-00-color": color,
	}
	return statusVals
}

func ContactFields(statusEntityID, ownerEntityID string, ownerEntityKey string) []entity.Field {
	nameField := entity.Field{
		Key:         schema.SeedFieldFNameKey,
		Name:        "first_name",
		DisplayName: "First Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "title"},
	}

	emailField := entity.Field{
		Key:         "uuid-00-email",
		Name:        "email",
		DisplayName: "Email",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{"layout": "sub-title"},
	}

	mobileField := entity.Field{
		Key:         "uuid-00-mobile-numbers",
		Name:        "mobile_numbers",
		DisplayName: "Mobile Numbers",
		DataType:    entity.TypeList,
		DomType:     entity.DomMultiSelect,
		Field: &entity.Field{
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
		DataType:    entity.TypeString,
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
	}

	statusField := entity.Field{
		Key:         "uuid-00-status",
		Name:        "status",
		DisplayName: "Status",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       statusEntityID,
		Meta:        map[string]string{"display_gex": "uuid-00-name", "verb": "uuid-00-verb"},
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
		Meta:        map[string]string{"display_gex": ownerEntityKey},
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
		"uuid-00-lf-stage":       "lead",
		"uuid-00-status":         []interface{}{statusID},
		"uuid-00-owner":          []interface{}{},
	}
	return contactVals
}

func CompanyFields() []entity.Field {
	nameField := entity.Field{
		Key:         "uuid-00-name",
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	websiteField := entity.Field{
		Key:         "uuid-00-website",
		Name:        "website",
		DisplayName: "Website",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	return []entity.Field{nameField, websiteField}
}

func CompanyVals(name, website string) map[string]interface{} {
	companyVals := map[string]interface{}{
		"uuid-00-name":    name,
		"uuid-00-website": website,
	}
	return companyVals
}

func TicketFields(statusEntityID string) []entity.Field {
	nameField := entity.Field{
		Key:         "uuid-00-subject",
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	statusField := entity.Field{
		Key:         "uuid-00-status",
		Name:        "status",
		DisplayName: "Status",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
		RefID:       statusEntityID,
		Meta:        map[string]string{"display_gex": "uuid-00-name"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{nameField, statusField}
}

func TicketVals(name, statusID string) map[string]interface{} {
	ticketVals := map[string]interface{}{
		"uuid-00-subject": name,
		"uuid-00-status":  []interface{}{statusID},
	}
	return ticketVals
}

func TaskFields(contactEntityID, statusEntityID string, stItem1, stItem2, stItem3 string) []entity.Field {
	descField := entity.Field{
		Key:         "uuid-00-desc",
		Name:        "desc",
		DisplayName: "Notes",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	dueByField := entity.Field{
		Key:         "uuid-00-due-by",
		Name:        "due_by",
		DisplayName: "Due By",
		DomType:     entity.DomText,
		DataType:    entity.TypeDataTime,
	}

	contactField := entity.Field{
		Key:         "uuid-00-contact",
		Name:        "contact",
		DisplayName: "Associated To",
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

	reminderField := entity.Field{
		Key:         "uuid-00-reminder",
		Name:        "reminder",
		DisplayName: "Reminder",
		DomType:     entity.DomText,
		DataType:    entity.TypeDataTime,
		ActionID:    contactField.Key,
	}

	statusField := entity.Field{
		Key:         "uuid-00-status",
		Name:        "status",
		DisplayName: "Status",
		DomType:     entity.DomAutoSelect,
		DataType:    entity.TypeReference,
		RefID:       statusEntityID,
		Meta:        map[string]string{"display_gex": "uuid-00-name", "verb": "uuid-00-verb", "layout": "verb"},
		Choices: []entity.Choice{
			{
				ID:         stItem1,
				Expression: fmt.Sprintf("{{%s.%s}} af {now}", "self", dueByField.Key),
			},
			{
				ID:         stItem3,
				Expression: fmt.Sprintf("{{%s.%s}} bf {now}", "self", dueByField.Key),
			},
		},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{descField, contactField, statusField, dueByField, reminderField}
}

func TaskVals(desc, contactID string) map[string]interface{} {
	taskVals := map[string]interface{}{
		"uuid-00-desc":     desc,
		"uuid-00-contact":  []interface{}{contactID},
		"uuid-00-status":   []interface{}{},
		"uuid-00-reminder": util.FormatTimeGo(time.Now()),
		"uuid-00-due-by":   util.FormatTimeGo(time.Now()),
	}
	return taskVals
}

func DealFields(contactEntityID string, flowEntityID, nodeEntityID string) []entity.Field {
	dealName := entity.Field{
		Key:         "uuid-00-deal-name",
		Name:        "deal_name",
		DisplayName: "Deal Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
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
		Meta:        map[string]string{"display_gex": "uuid-00-fname"},
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
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       nodeEntityID,
		Dependent: &entity.Dependent{
			ParentKey:    pipeField.Key,
			ReferenceKey: "flow_id",
		},
		Meta: map[string]string{"display_gex": "uuid-00-fname", "node": "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      pipeField.Key,
			Value:    "--",
		},
	}

	return []entity.Field{dealName, dealAmount, contactsField, pipeField, pipeStageField}
}

func DealVals(name string, amount int, contactID1, contactID2, flowID string) map[string]interface{} {
	dealVals := map[string]interface{}{
		"uuid-00-deal-name":   name,
		"uuid-00-deal-amount": amount,
		"uuid-00-contacts":    []interface{}{contactID1, contactID2},
		"uuid-00-pipe":        []interface{}{flowID},
		"uuid-00-pipe-stage":  []interface{}{},
	}
	return dealVals
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
		DataType: entity.TypeDataTime,
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
