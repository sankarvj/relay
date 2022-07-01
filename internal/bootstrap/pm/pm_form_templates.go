package pm

import (
	"fmt"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

func taskTemplates(desc string, thisEntity entity.Entity, actorEntity entity.Entity) map[string]interface{} {
	taskVals := make(map[string]interface{}, 0)
	namedFieldsMap := entity.NamedFieldsObjMap(thisEntity.FieldsIgnoreError())

	for name, f := range namedFieldsMap {
		if f.IsTitleLayout() {
			taskVals[f.Key] = fmt.Sprintf("%s", desc)
		}
		switch name {
		case "reminder", "due_by":
			taskVals[f.Key] = util.FormatTimeGo(time.Now())
		}
	}

	return taskVals
}

func inviteTemplates(desc string, thisEntity entity.Entity, actorEntity entity.Entity) map[string]interface{} {
	inviteVals := make(map[string]interface{}, 0)
	namedFieldsMap := entity.NamedFieldsObjMap(thisEntity.FieldsIgnoreError())

	var associatedContactKey string
	for _, f := range actorEntity.FieldsIgnoreError() {
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
