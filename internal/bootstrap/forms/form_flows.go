package forms

import (
	"github.com/google/uuid"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func FlowFields() []entity.Field {
	actualFlowIDFieldID := uuid.New().String()
	actualFlowIDField := entity.Field{
		Key:         actualFlowIDFieldID,
		Name:        "flow_id",
		DisplayName: "ID",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
	}

	actualFlowNameFieldID := uuid.New().String()
	actualFlowNameField := entity.Field{
		Key:         actualFlowNameFieldID,
		Name:        "flow_name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	return []entity.Field{actualFlowIDField, actualFlowNameField}
}

func NodeFields() []entity.Field {
	actualFlowIDFieldID := uuid.New().String()
	actualFlowIDFIeld := entity.Field{
		Key:         actualFlowIDFieldID,
		Name:        "node_id",
		DisplayName: "ID",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
	}

	return []entity.Field{actualFlowIDFIeld}
}
