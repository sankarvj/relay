package entity

import (
	"encoding/json"
	"log"

	"github.com/pkg/errors"
)

// Field represents structural format of attributes in entity
type Field struct {
	Name        string            `json:"name" validate:"required"`
	DisplayName string            `json:"display_name" validate:"required"` //do we need this? why not use name for display
	Key         string            `json:"key" validate:"required"`
	Value       interface{}       `json:"value" validate:"required"`
	DataType    DType             `json:"data_type" validate:"required"`
	DomType     Dom               `json:"dom_type" validate:"required"`
	Field       *Field            `json:"field"`
	Meta        map[string]string `json:"meta"` //shall we move the extra prop to this meta or shall we keep it flat?
	Choices     []Choice          `json:"choices"`
	RefID       string            `json:"ref_id"`    // this could be another entity_id for reference, pipeline_id for odd with pipleline/playbook
	ActionID    string            `json:"action_id"` // another field_id for datetime with reminder/dueby
}

type FieldMeta struct {
	Unique     string `json:"unique"`
	Mandatory  string `json:"mandatory"`
	Hidden     string `json:"hidden"`
	Config     string `json:"config"`     //UI property useful only during display
	Expression string `json:"expression"` //expression is a double purpose property - executes the checks like, field.value > 100 < 200 or field.value == 'vijay' during "save", checks the operator during segmenting
	Link       string `json:"link"`       //useful for autocomplete. If number of choices greater than 100
	DisplayGex string `json:"display_gex"`
	Layout     string `json:"layout"`
}

type Choice struct {
	ID           string      `json:"id"`
	DisplayValue interface{} `json:"display_value"`
	Expression   string      `json:"expression"`
}

//DType defines the data type of field
type DType string

//Mode for the entity spcifies certain entity specific characteristics
//Keep this as minimal and add a sub-type for data types such as decimal,boolean,time & date
const (
	TypeString    DType = "S"
	TypeNumber          = "N"
	TypeDataTime        = "T"
	TypeList            = "L"
	TypeReference       = "R"
	TypeOdd             = "O"
)

//Dom defines the visual representation of the field
type Dom string

//const defines the types of visual representation dom
const (
	DomText          Dom = "TE"
	DomTextArea          = "TA"
	DomStatus            = "ST"
	DomAutoComplete      = "AC"
	DomSelect            = "SE" // the default dom for the reference field units. This type mandates the choices limit to 20
	DomAutoSelect        = "AS" // same as select but with the twist for auto fill. refer status
	DomDate              = "DA"
	DomTime              = "TI"
	DomMinute            = "MI"
	DomMultiSelect       = "MS"
	DomNotApplicable     = "NA" // the dom for the reference field with no UI needed
	DomPipeline          = "PL"
	DomPlayBook          = "PB"
	DomReminder          = "RE"
)

//field_unit expression
const (
	FuExpNone   = "none"
	FuExpPos    = "pos"    //set this on positive expression of due_by
	FuExpNeg    = "neg"    //set this on negative expression of the due_by
	FuExpManual = "manual" //keep as it is unless manually changes
)

//ValueAddFields updates the values of entity fields along with the config
func ValueAddFields(entityFields []Field, itemFields map[string]interface{}) []Field {
	valueAddedFields := make([]Field, 0)
	for _, field := range entityFields {
		if val, ok := itemFields[field.Key]; ok {
			field.Value = val
		}
		valueAddedFields = append(valueAddedFields, field)
	}
	return valueAddedFields
}

// Fields parses attribures to fields
func (e Entity) Fields() ([]Field, error) {
	fields, err := unmarshalFields(e.Fieldsb)
	if err != nil {
		err = errors.Wrapf(err, "error while unmarshalling entity attributes to fields type %q", e.ID)
		log.Println(err)
		return make([]Field, 0), err
	}
	return fields, nil
}

func (e Entity) Key(name string) string {
	fields, err := e.Fields()
	if err != nil {
		return ""
	}
	return NamedKeysMap(fields)[name]
}

func (f *Field) SetDisplayGex(key string) {
	if f.Meta == nil {
		f.Meta = make(map[string]string, 0)
	}
	f.Meta["display_gex"] = key
}

func (f Field) isConfig() bool {
	if val, ok := f.Meta["config"]; ok && val == "true" {
		return true
	}
	return false
}

func (f Field) IsReference() bool {
	if f.DataType == TypeReference {
		return true
	}
	return false
}

func (f Field) IsPipe() bool {
	if f.DomType == DomPipeline || f.DomType == DomPlayBook {
		return true
	}
	return false
}

func (f Field) IsNotApplicable() bool {
	if f.DomType == DomNotApplicable {
		return true
	}
	return false
}

func (f Field) DisplayGex() string {
	if val, ok := f.Meta["display_gex"]; ok {
		return val
	}
	return ""
}

func refFields(fields []Field) map[string]string {
	referenceFieldsMap := make(map[string]string, 0)
	for _, f := range fields {
		if f.IsReference() { // TODO: also check if customer explicitly asks for it. Don't do this for all the reference fields
			referenceFieldsMap[f.Key] = f.RefID
		}
	}
	return referenceFieldsMap
}

func NamedKeysMap(entityFields []Field) map[string]string {
	params := map[string]string{}
	for _, field := range entityFields {
		params[field.Name] = field.Key
	}
	return params
}

func unmarshalFields(fieldsB string) ([]Field, error) {
	var fields []Field
	if err := json.Unmarshal([]byte(fieldsB), &fields); err != nil {
		return nil, err
	}
	return fields, nil
}
