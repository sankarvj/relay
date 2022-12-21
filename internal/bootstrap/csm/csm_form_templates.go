package csm

import (
	"fmt"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

func taskTemplates(name, desc, statusID string, thisEntity entity.Entity) map[string]interface{} {
	taskVals := make(map[string]interface{}, 0)
	namedFieldsMap := entity.NameMap(thisEntity.EasyFields())

	for nameOfField, f := range namedFieldsMap {
		if f.IsTitleLayout() {
			taskVals[f.Key] = name
		}
		switch nameOfField {
		case "name":
			taskVals[f.Key] = name
		case "desc":
			taskVals[f.Key] = desc
		case "status":
			taskVals[f.Key] = []interface{}{statusID}
		// case "associated_contacts":
		// 	taskVals[f.Key] = []interface{}{contactID}
		case "reminder", "due_by":
			taskVals[f.Key] = util.FormatTimeGo(time.Now())
		}
	}

	return taskVals
}

func inviteTemplates(desc string, thisEntity entity.Entity, actorEntity entity.Entity) map[string]interface{} {
	inviteVals := make(map[string]interface{}, 0)
	namedFieldsMap := entity.NameMap(thisEntity.EasyFields())

	var associatedContactKey string
	for _, f := range actorEntity.EasyFields() {
		if f.Name == "contact" {
			associatedContactKey = f.Key
		}
	}

	for name, f := range namedFieldsMap {

		switch name {
		case "email":
			inviteVals[f.Key] = fmt.Sprintf("{{%s.%s.email}}", actorEntity.ID, associatedContactKey)
		case "role":
			inviteVals[f.Key] = []interface{}{"VISITOR"}
		case "body":
			inviteVals[f.Key] = desc
		}
	}
	return inviteVals
}

func dealTemplates(thisEntity entity.Entity, actorEntity entity.Entity, flowID string) (map[string]interface{}, map[string]interface{}) {
	dealMeta := make(map[string]interface{}, 0)
	dealVals := make(map[string]interface{}, 0)
	namedFieldsMap := entity.NameMap(thisEntity.EasyFields())

	var titleKey string
	for _, f := range actorEntity.EasyFields() {
		if f.IsTitleLayout() {
			titleKey = f.Key
		}
	}

	for name, f := range namedFieldsMap {
		if f.IsTitleLayout() {
			atmention := `<p><span class=\"mention\" data-index=\"0\" data-denotation-char=\"#\" data-id=\"{{%s.%s}}\" data-value=\"Name\"><span contenteditable=\"false\"><span class=\"ql-mention-denotation-char\">#</span>Name</span></span>'s Base Deal</p>`
			dealVals[f.Key] = fmt.Sprintf(atmention, actorEntity.ID, titleKey)
		}
		switch name {
		case "pipeline":
			dealVals[f.Key] = []interface{}{flowID}
		case "close_date":
			dealVals[f.Key] = util.FormatTimeGo(time.Now())
		case "associated_companies":
			dealVals[f.Key] = []interface{}{fmt.Sprintf("{{%s.id}}", actorEntity.ID)}
		}

	}
	return dealVals, dealMeta
}

func contactTemplates(thisEntity entity.Entity, actorEntity entity.Entity, leadStatus string) (map[string]interface{}, map[string]interface{}) {
	contactMeta := make(map[string]interface{}, 0)
	contactVals := make(map[string]interface{}, 0)
	namedFieldsMap := entity.NameMap(thisEntity.EasyFields())

	var dealAmountField entity.Field
	for _, f := range actorEntity.EasyFields() {
		if f.Name == "deal_amount" {
			dealAmountField = f
		}
	}

	for name, f := range namedFieldsMap {
		switch name {
		case "lead_status":
			contactVals[f.Key] = []interface{}{leadStatus}
		case "total_revenue":
			val := fmt.Sprintf("{{%s.%s}}", actorEntity.ID, dealAmountField.Key)
			contactVals[f.Key] = val
			contactMeta[val] = dealAmountField.DisplayName
		case "tags":
			contactVals[f.Key] = []interface{}{"Enterprise customer"}
		}
	}

	return contactVals, contactMeta
}
