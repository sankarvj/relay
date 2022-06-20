package em

import (
	"fmt"
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
			dealVals[f.Key] = fmt.Sprintf("{{%s.%s}}'s Deal", actorEntity.ID, titleKey)
		}
		switch name {
		case "pipeline":
			dealVals[f.Key] = []interface{}{flowID}
		case "close_date":
			dealVals[f.Key] = util.FormatTimeGo(time.Now())
		case "company":
			dealVals[f.Key] = []interface{}{fmt.Sprintf("{{%s.id}}", actorEntity.ID)}
		}

	}
	return dealVals
}

func taskTemplates(thisEntity entity.Entity, actorEntity entity.Entity) map[string]interface{} {
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
			taskVals[f.Key] = fmt.Sprintf("Schedule a call for {{%s.%s}}", actorEntity.ID, titleKey)
		}
		switch name {
		case "reminder", "due_by":
			taskVals[f.Key] = util.FormatTimeGo(time.Now())
		}
	}

	return taskVals
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
