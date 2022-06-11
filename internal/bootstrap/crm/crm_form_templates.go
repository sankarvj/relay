package crm

import (
	"fmt"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

func dealTemplates(actorEntity entity.Entity, flowID string) map[string]interface{} {
	var titleKey string
	for _, f := range actorEntity.FieldsIgnoreError() {
		if f.IsTitleLayout() {
			titleKey = f.Key
		}
	}

	dealVals := map[string]interface{}{
		"uuid-00-deal-name":  fmt.Sprintf("{{%s.%s}}'s Deal", actorEntity.ID, titleKey),
		"uuid-00-contacts":   []interface{}{},
		"uuid-00-pipe":       []interface{}{flowID},
		"uuid-00-pipe-stage": []interface{}{},
		"uuid-00-company":    []interface{}{fmt.Sprintf("{{%s.id}}", actorEntity.ID)},
		"uuid-00-close-date": util.FormatTimeGo(time.Now()),
	}
	return dealVals
}

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
		"uuid-00-deal":          []interface{}{fmt.Sprintf("{{%s.%s}}", actorEntity.ID, "id")},
		"uuid-00-reminder":      util.FormatTimeGo(time.Now()),
		"uuid-00-due-by":        util.FormatTimeGo(time.Now()),
		"uuid-00-type":          []interface{}{},
		"uuid-00-mail-template": []interface{}{},
		"uuid-00-pipe-stage":    []interface{}{},
	}
	return taskVals
}

func ticketTemplates(actorEntity entity.Entity) map[string]interface{} {
	var titleKey string
	for _, f := range actorEntity.FieldsIgnoreError() {
		if f.IsTitleLayout() {
			titleKey = f.Key
		}
	}

	ticketVals := map[string]interface{}{
		"uuid-00-subject": fmt.Sprintf("Please prepare the invoice for the deal {{%s.%s}}", actorEntity.ID, titleKey),
		"uuid-00-contact": []interface{}{},
		"uuid-00-company": []interface{}{},
		"uuid-00-deal":    []interface{}{fmt.Sprintf("{{%s.%s}}", actorEntity.ID, "id")},
		"uuid-00-status":  []interface{}{},
	}
	return ticketVals
}
