package csm

import (
	"fmt"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

func taskTemplates(actorEntity entity.Entity) map[string]interface{} {
	var titleKey string
	for _, f := range actorEntity.FieldsIgnoreError() {
		if f.IsTitleLayout() {
			titleKey = f.Key
		}
	}

	taskVals := map[string]interface{}{
		"uuid-00-desc":          fmt.Sprintf("Schedule a call for {{%s.%s}}", actorEntity.ID, titleKey),
		"uuid-00-contact":       []interface{}{},
		"uuid-00-status":        []interface{}{},
		"uuid-00-company":       []interface{}{},
		"uuid-00-project":       []interface{}{fmt.Sprintf("{{%s.%s}}", actorEntity.ID, "id")},
		"uuid-00-reminder":      util.FormatTimeGo(time.Now()),
		"uuid-00-due-by":        util.FormatTimeGo(time.Now()),
		"uuid-00-type":          []interface{}{},
		"uuid-00-mail-template": []interface{}{},
		"uuid-00-pipe-stage":    []interface{}{},
	}
	return taskVals
}
