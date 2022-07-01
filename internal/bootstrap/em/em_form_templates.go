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

func taskTemplates(desc string, thisEntity entity.Entity, actorEntity entity.Entity) map[string]interface{} {
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
			taskVals[f.Key] = fmt.Sprintf("%s {{%s.%s}}", desc, actorEntity.ID, titleKey)
		}
		switch name {
		case "reminder", "due_by":
			taskVals[f.Key] = util.FormatTimeGo(time.Now())
		}
	}

	return taskVals
}

func assetRequestTemplates(desc, assetID, statusID string, thisEntity entity.Entity) map[string]interface{} {
	arVals := make(map[string]interface{}, 0)
	namedFieldsMap := entity.NamedFieldsObjMap(thisEntity.FieldsIgnoreError())

	for name, f := range namedFieldsMap {
		switch name {
		case "comments":
			arVals[f.Key] = fmt.Sprintf("%s", desc)
		case "asset":
			arVals[f.Key] = []interface{}{assetID}
		case "request_status":
			arVals[f.Key] = []interface{}{statusID}
		}
	}
	return arVals
}

func serviceRequestTemplates(desc, serviceID, statusID string, thisEntity entity.Entity) map[string]interface{} {
	arVals := make(map[string]interface{}, 0)
	namedFieldsMap := entity.NamedFieldsObjMap(thisEntity.FieldsIgnoreError())

	for name, f := range namedFieldsMap {
		switch name {
		case "desc":
			arVals[f.Key] = fmt.Sprintf("%s", desc)
		case "service":
			arVals[f.Key] = []interface{}{serviceID}
		case "request_status":
			arVals[f.Key] = []interface{}{statusID}
		}
	}
	return arVals
}

func inviteTemplates(desc string, thisEntity entity.Entity, actorEntity entity.Entity) map[string]interface{} {
	inviteVals := make(map[string]interface{}, 0)
	namedFieldsMap := entity.NamedFieldsObjMap(thisEntity.FieldsIgnoreError())

	var emailKey string
	for _, f := range actorEntity.FieldsIgnoreError() {
		if f.IsEmail() {
			emailKey = f.Key
		}
	}

	for name, f := range namedFieldsMap {
		switch name {
		case "email":
			inviteVals[f.Key] = fmt.Sprintf("{{%s.%s}}", actorEntity.ID, emailKey)
		case "role":
			inviteVals[f.Key] = []interface{}{"USER"}
		case "body":
			inviteVals[f.Key] = desc
		}
	}
	return inviteVals
}
