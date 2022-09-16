package crm

import (
	"fmt"
	"log"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

func dealTemplates(thisEntity entity.Entity, actorEntity entity.Entity, flowID string) map[string]interface{} {
	dealVals := make(map[string]interface{}, 0)
	namedFieldsMap := entity.NamedFieldsObjMap(thisEntity.FieldsIgnoreError())

	var titleKey string
	for _, f := range actorEntity.FieldsIgnoreError() {
		if f.IsTitleLayout() {
			titleKey = f.Key
		}
	}

	for name, f := range namedFieldsMap {
		if f.IsTitleLayout() {
			atmention := `<p><span class="mention" data-index="0" data-denotation-char="#" data-id="{{%s.%s}}" data-value="Name"><span contenteditable="false"><span class="ql-mention-denotation-char">#</span>Name</span></span>'s Deal</p>`
			log.Println("atmention", fmt.Sprintf(atmention, actorEntity.ID, titleKey)) // not working, hence commented
			dealVals[f.Key] = "Base Deal"
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
	return dealVals
}

func taskTemplates(msg string, thisEntity entity.Entity, actorEntity entity.Entity, withToken bool) map[string]interface{} {
	taskVals := make(map[string]interface{}, 0)
	namedFieldsMap := entity.NamedFieldsObjMap(thisEntity.FieldsIgnoreError())

	var titleKey string
	for _, f := range actorEntity.FieldsIgnoreError() {
		if f.IsTitleLayout() {
			titleKey = f.Key
		}
	}

	for name, f := range namedFieldsMap {
		if f.IsTitleLayout() {
			if withToken {
				taskVals[f.Key] = fmt.Sprintf("%s {{%s.%s}}", msg, actorEntity.ID, titleKey)
			} else {
				taskVals[f.Key] = msg
			}

		}
		switch name {
		case "reminder":
			taskVals[f.Key] = util.FormatTimeGo(time.Now().Add(time.Hour * 24 * time.Duration(2)))
		case "due_by":
			taskVals[f.Key] = util.FormatTimeGo(time.Now().Add(time.Hour * 24 * time.Duration(3)))
		}
	}

	return taskVals
}

func contactTemplates(thisEntity entity.Entity, actorEntity entity.Entity, leadStatus string) map[string]interface{} {
	contactVals := make(map[string]interface{}, 0)
	namedFieldsMap := entity.NamedFieldsObjMap(thisEntity.FieldsIgnoreError())

	var dealAmountKey string
	for _, f := range actorEntity.FieldsIgnoreError() {
		if f.Name == "deal_amount" {
			dealAmountKey = f.Key
		}
	}

	for name, f := range namedFieldsMap {
		switch name {
		case "lead_status":
			contactVals[f.Key] = []interface{}{leadStatus}
		case "total_revenue":
			contactVals[f.Key] = fmt.Sprintf("{{%s.%s}}", actorEntity.ID, dealAmountKey)
		case "tags":
			contactVals[f.Key] = []interface{}{"Enterprise customer"}
		}
	}

	return contactVals
}

func ticketTemplates(thisEntity entity.Entity, actorEntity entity.Entity) map[string]interface{} {
	ticketVals := make(map[string]interface{}, 0)
	namedFieldsMap := entity.NamedFieldsObjMap(thisEntity.FieldsIgnoreError())

	var titleKey string
	for _, f := range actorEntity.FieldsIgnoreError() {
		if f.IsTitleLayout() {
			titleKey = f.Key
		}
	}

	for _, f := range namedFieldsMap {
		if f.IsTitleLayout() {
			ticketVals[f.Key] = fmt.Sprintf("Please prepare the invoice for the deal {{%s.%s}}", actorEntity.ID, titleKey)
		}
	}

	return ticketVals
}
