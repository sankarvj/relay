package graphdb

import (
	"encoding/json"

	"github.com/pkg/errors"
)

//DType defines the data type of field
type DType string

//Mode for the entity spcifies certain entity specific characteristics
const (
	TypeString    DType = "S"
	TypeNumber          = "N"
	TypeDataTime        = "T"
	TypeList            = "L"
	TypeReference       = "R"
)

const (
	FieldIdKey = "id"
)

//Field is the subset of entity field. Make sure it is on par with Entity Field
type Field struct {
	Key        string      `json:"key" validate:"required"`
	Value      interface{} `json:"value" validate:"required"`
	DataType   DType       `json:"data_type" validate:"required"`
	Expression string      `json:"expression"`
	Field      *Field      `json:"field"`
}

func Fields(jsonB string) ([]Field, error) {
	var fields []Field
	if err := json.Unmarshal([]byte(jsonB), &fields); err != nil {
		return nil, errors.Wrapf(err, "error while unmarshalling segment attributes to fields")
	}
	return fields, nil
}

func FillFieldValues(eFields []Field, itemProps map[string]interface{}) []Field {
	uFields := make([]Field, 0)
	for _, field := range eFields {
		if val, ok := itemProps[field.Key]; ok {
			field.Value = val
		}
		uFields = append(uFields, field)
	}
	return uFields
}

func (f Field) IsKeyId() bool {
	return f.Key == FieldIdKey
}

func (f Field) RefList() []map[string]string {
	return f.Value.([]map[string]string)
}

func (f Field) SetRef(entityID, itemID string) Field {
	f.Value = RefMap(entityID, itemID)
	return f
}

func fetchRef(ref map[string]string) (string, string) {
	return ref["entity_id"], ref["item_id"]
}

func RefMap(entityID, itemID string) []map[string]string {
	return []map[string]string{
		{
			"entity_id": entityID,
			"item_id":   itemID,
		},
	}
}
