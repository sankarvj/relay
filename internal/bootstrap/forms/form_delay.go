package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func DelayBy2Day(titleFieldKey, delayByFieldKey string) map[string]interface{} {
	delayVals := map[string]interface{}{
		titleFieldKey:   "2 day delay",
		delayByFieldKey: 2880,
	}
	return delayVals
}

func DelayBy1Day(titleFieldKey, delayByFieldKey string) map[string]interface{} {
	delayVals := map[string]interface{}{
		titleFieldKey:   "1 day delay",
		delayByFieldKey: 1440,
	}
	return delayVals
}

func DelayBy5Hr(titleFieldKey, delayByFieldKey string) map[string]interface{} {
	delayVals := map[string]interface{}{
		titleFieldKey:   "5 hr delay",
		delayByFieldKey: 300,
	}
	return delayVals
}

func DelayBy1Hr(titleFieldKey, delayByFieldKey string) map[string]interface{} {
	delayVals := map[string]interface{}{
		titleFieldKey:   "1 hr delay",
		delayByFieldKey: 60,
	}
	return delayVals
}

func DelayBy30Min(titleFieldKey, delayByFieldKey string) map[string]interface{} {
	delayVals := map[string]interface{}{
		titleFieldKey:   "30 min delay",
		delayByFieldKey: 30,
	}
	return delayVals
}

func DelayBy5Min(titleFieldKey, delayByFieldKey string) map[string]interface{} {
	delayVals := map[string]interface{}{
		titleFieldKey:   "5 min delay",
		delayByFieldKey: 5,
	}
	return delayVals
}

func DelayFields() []entity.Field {
	titleFieldID := uuid.New().String()
	titleField := entity.Field{
		Key:         titleFieldID,
		Name:        "title",
		DisplayName: "Title",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
	}

	delayByFieldID := uuid.New().String()
	delayByField := entity.Field{
		Key:         delayByFieldID,
		Name:        "delay_by",
		DisplayName: "Delay in mins",
		DomType:     entity.DomText,
		DataType:    entity.TypeNumber,
	}

	return []entity.Field{titleField, delayByField}
}
