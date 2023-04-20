package incident

import (
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
