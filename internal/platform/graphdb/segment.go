package graphdb

import (
	"encoding/json"

	"github.com/pkg/errors"
)

//DType defines the data type of field
type DType string

//Mode for the entity spcifies certain entity specific characteristics
//Keep this as minimal and add a sub-type for data types such as decimal,boolean,time & date
const (
	TypeString         DType = "S"
	TypeNumber               = "N"
	TypeDateTime             = "T"
	TypeDateRange            = "TR"
	TypeDateTimeMillis       = "TM"
	TypeList                 = "L"
	TypeReference            = "R"
	TypeWist                 = "W" // wist( implies: where clause list) is handled as `Where clause IN` instead of `Has/Contains` in list & reference
)

const (
	FieldIdKey = "id"
)

//Field is the subset of entity field. Make sure it is on par with Entity Field
type Field struct {
	Key          string      `json:"key" validate:"required"`
	Value        interface{} `json:"value" validate:"required"`
	Min          interface{} `json:"min"` //min date
	Max          interface{} `json:"max"` //max date
	DataType     DType       `json:"data_type" validate:"required"`
	Expression   string      `json:"expression"`
	RefID        string      `json:"ref_id"`
	Field        *Field      `json:"field"`
	UnlinkOffset int         `json:"unlink_offset"` // 0 if nothing to delete
	Aggr         string      `json:"aggr"`
	WithAlias    string      `json:"with_alias"`
	IsReverse    bool        `json:"is_reverse"`
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
