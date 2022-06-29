package base

import (
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func APIFields() []entity.Field {
	path := entity.Field{
		Key:      "uuid-00-path",
		Name:     "path",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "/actuator/info",
	}

	host := entity.Field{
		Key:      "uuid-00-host",
		Name:     "host",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "https://stage.freshcontacts.io",
	}

	method := entity.Field{
		Key:      "uuid-00-method",
		Name:     "method",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "GET",
	}

	headers := entity.Field{
		Key:      "uuid-00-headers",
		Name:     "headers",
		DomType:  entity.DomText,
		DataType: entity.TypeString,
		Value:    "{\"X-ClientToken\":\"mcr eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjExNTc2NTAwMjk2fQ.1KtXw_YgxbJW8ibv_v2hfpInjQKC6enCh9IO1ziV2RA\"}",
	}

	return []entity.Field{path, host, method, headers}
}

func DelayVals() map[string]interface{} {
	delayVals := map[string]interface{}{
		"uuid-00-title":    "1 hr delay",
		"uuid-00-delay-by": 1,
		"uuid-00-repeat":   "false",
	}
	return delayVals
}

func DelayFields() []entity.Field {
	titleField := entity.Field{
		Key:         "uuid-00-title",
		Name:        "title",
		DisplayName: "Title",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: "title"},
	}

	delay_by := entity.Field{
		Key:         "uuid-00-delay-by",
		Name:        "delay_by",
		DisplayName: "Delay By",
		DomType:     entity.DomText,
		DataType:    entity.TypeNumber,
	}

	repeat := entity.Field{
		Key:         "uuid-00-repeat",
		Name:        "repeat",
		DisplayName: "Repeat",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Value:       "true",
	}

	return []entity.Field{titleField, delay_by, repeat}
}
