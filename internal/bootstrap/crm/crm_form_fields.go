package crm

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/reference"
)

func TaskVals(actorEntity entity.Entity, desc, contactID string) map[string]interface{} {

	taskVals := make(map[string]interface{}, 0)
	namedFieldsMap := entity.NamedFieldsObjMap(actorEntity.FieldsIgnoreError())

	for name, f := range namedFieldsMap {
		switch name {
		case "desc":
			taskVals[f.Key] = desc
		case "contact":
			taskVals[f.Key] = []interface{}{contactID}
		case "reminder", "due_by":
			taskVals[f.Key] = util.FormatTimeGo(time.Now())
		}

	}
	return taskVals
}

func MeetingFields(contactEntityID, companyEntityID, dealEntityID string, contactEntityEmailFieldID, contactEntityFirstNameFieldID, companyEntityNameFieldID, dealEntityNameFieldID string) []entity.Field {

	titleFieldID := uuid.New().String()
	titleField := entity.Field{
		Key:         titleFieldID,
		Name:        "cal_title",
		DisplayName: "Title",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	summaryFieldID := uuid.New().String()
	summaryField := entity.Field{
		Key:         summaryFieldID,
		Name:        "summary",
		DisplayName: "Summary",
		DomType:     entity.DomTextArea,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle},
	}

	attendessFieldID := uuid.New().String()
	attendessField := entity.Field{
		Key:         attendessFieldID,
		Name:        "attendess",
		DisplayName: "Attendess",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       contactEntityID,
		Who:         entity.WhoAssignee,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: contactEntityEmailFieldID, entity.MetaKeyLayout: entity.MetaLayoutUsers},
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
		Meta:        map[string]string{entity.MetaKeyRow: "true"},
		Who:         entity.WhoEndTime,
	}

	timezoneFieldID := uuid.New().String()
	timezoneField := entity.Field{
		Key:         timezoneFieldID,
		Name:        "timezone",
		DisplayName: "Timezone",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyHidden: "true"},
	}

	createdAtFieldID := uuid.New().String()
	createdAtField := entity.Field{
		Key:         createdAtFieldID,
		Name:        "created_at",
		DisplayName: "Created At",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
		Meta:        map[string]string{entity.MetaKeyHidden: "true"},
	}

	updatedAtFieldID := uuid.New().String()
	updatedAtField := entity.Field{
		Key:         updatedAtFieldID,
		Name:        "updated_at",
		DisplayName: "Updated At",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
		Meta:        map[string]string{entity.MetaKeyHidden: "true"},
	}

	contactFieldID := uuid.New().String()
	contactField := entity.Field{
		Key:         contactFieldID,
		Name:        "associated_contacts",
		DisplayName: "Associated Contact",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       contactEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: contactEntityFirstNameFieldID},
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
		DisplayName: "Associated Company",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       companyEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: companyEntityNameFieldID, entity.MetaKeyRow: "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	dealFieldID := uuid.New().String()
	dealField := entity.Field{
		Key:         dealFieldID,
		Name:        "associated_deals",
		DisplayName: "Associated Deal",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       dealEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: dealEntityNameFieldID, entity.MetaKeyRow: "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{titleField, summaryField, attendessField, startTimeField, endTimeField, timezoneField, createdAtField, updatedAtField, contactField, companyField, dealField}
}

func DealFields(contactEntityID, contactEntityKey, companyEntityID, companyEntityKey string, flowEntityID, nodeEntityID, nodeKey string) []entity.Field {
	dealNameFieldID := uuid.New().String()
	dealNameField := entity.Field{
		Key:         dealNameFieldID,
		Name:        "deal_name",
		DisplayName: "Deal Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	dealAmountFieldID := uuid.New().String()
	dealAmountField := entity.Field{
		Key:         dealAmountFieldID,
		Name:        "deal_amount",
		DisplayName: "Deal Amount",
		DomType:     entity.DomText,
		DataType:    entity.TypeNumber,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle},
	}

	contactsFieldID := uuid.New().String()
	contactsField := entity.Field{
		Key:         contactsFieldID,
		Name:        "associated_contacts",
		DisplayName: "Associated Contacts",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       contactEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: contactEntityKey, entity.MetaMultiChoice: "true", entity.MetaKeyLayout: entity.MetaLayoutUsers},
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
		Meta:        map[string]string{entity.MetaKeyDisplayGex: companyEntityKey, entity.MetaMultiChoice: "true"},
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

	closeDateFieldID := uuid.New().String()
	closeDateField := entity.Field{
		Key:         closeDateFieldID,
		Name:        "close_date",
		DisplayName: "Close date",
		DomType:     entity.DomText,
		DataType:    entity.TypeDate,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutDate},
	}

	return []entity.Field{dealNameField, dealAmountField, contactsField, companyField, pipeField, pipeStageField, closeDateField}
}

func DealVals(dealEntity entity.Entity, name string, amount int, contactID1, contactID2, flowID string) map[string]interface{} {
	dealVals := map[string]interface{}{
		"deal_name":            name,
		"deal_amount":          amount,
		"associated_contacts":  []interface{}{contactID1, contactID2},
		"associated_companies": []interface{}{},
		"pipeline":             []interface{}{flowID},
		"pipeline_stage":       []interface{}{},
		"close_date":           util.FormatTimeGo(time.Now()),
	}
	return forms.KeyMap(dealEntity.NamedKeys(), dealVals)
}

func NoteFields(contactEntityID, contactEntityKey, companyEntityID, companyEntityKey, dealEntityID, dealEntityKey string) []entity.Field {
	descFieldID := uuid.New().String()
	descField := entity.Field{
		Key:         descFieldID,
		Name:        "desc",
		DisplayName: "Notes",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	contactFieldID := uuid.New().String()
	contactField := entity.Field{
		Key:         contactFieldID,
		Name:        "associated_contacts",
		DisplayName: "Associated To",
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
		DisplayName: "Associated To",
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

	dealFieldID := uuid.New().String()
	dealField := entity.Field{
		Key:         dealFieldID,
		Name:        "associated_deals",
		DisplayName: "Associated To",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       dealEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: dealEntityKey},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	return []entity.Field{descField, contactField, companyField, dealField}
}

func NoteVals(noteEntity entity.Entity, desc, contactID string) map[string]interface{} {
	noteVals := map[string]interface{}{
		"desc":                desc,
		"associated_contacts": []interface{}{contactID},
	}
	return forms.KeyMap(noteEntity.NamedKeys(), noteVals)
}

func TicketFields(contactEntityID, contactEntityKey, companyEntityID, companyEntityKey, statusEntityID, statusEntityKey string) []entity.Field {
	nameFieldID := uuid.New().String()
	nameField := entity.Field{
		Key:         nameFieldID,
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
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
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
	}

	contactFieldID := uuid.New().String()
	contactField := entity.Field{
		Key:         contactFieldID,
		Name:        "associated_contacts",
		DisplayName: "Associated Contacts",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       contactEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: contactEntityKey, entity.MetaKeyLayout: entity.MetaLayoutUsers},
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

	return []entity.Field{nameField, statusField, contactField, companyField}
}

func TicketVals(ticketEntity entity.Entity, name, statusID string) map[string]interface{} {
	ticketVals := map[string]interface{}{
		"name":                 name,
		"associated_contacts":  []interface{}{},
		"associated_companies": []interface{}{},
		"status":               []interface{}{statusID},
	}
	return forms.KeyMap(ticketEntity.NamedKeys(), ticketVals)
}
