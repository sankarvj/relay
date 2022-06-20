package forms

import "gitlab.com/vjsideprojects/relay/internal/entity"

func FlowFields() []entity.Field {
	actualFlowID := entity.Field{
		Key:         "uuid-00-flow-id",
		Name:        "flow_id",
		DisplayName: "ID",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
	}

	actualFlowName := entity.Field{
		Key:         "uuid-00-flow-name",
		Name:        "flow_name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
	}

	return []entity.Field{actualFlowID, actualFlowName}
}

func NodeFields() []entity.Field {
	actualFlowID := entity.Field{
		Key:         "uuid-00-node-id",
		Name:        "node_id",
		DisplayName: "ID",
		DomType:     entity.DomSelect,
		DataType:    entity.TypeReference,
	}

	return []entity.Field{actualFlowID}
}
