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
	Key          string      `json:"key" validate:"required"`
	Value        interface{} `json:"value" validate:"required"`
	DataType     DType       `json:"data_type" validate:"required"`
	Expression   string      `json:"expression"`
	RefID        string      `json:"ref_id"`
	Field        *Field      `json:"field"`
	UnlinkOffset int         `json:"unlink_offset"` // 0 if nothing to delete
	Aggr         string      `json:"aggr"`
	WithAlias    string      `json:"with_alias"`
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

func (f Field) doUnlink(index int) bool {
	return f.UnlinkOffset != 0 && index+1 >= f.UnlinkOffset
}

func (f Field) RefList() []map[string]string {
	if f.Value == nil {
		return []map[string]string{}
	}
	return f.Value.([]map[string]string)
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
