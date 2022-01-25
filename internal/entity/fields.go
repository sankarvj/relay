package entity

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/relationship"
)

const (
	Verb    = "verb"
	VerbKey = "uuid-00-verb"
)

//DType defines the data type of field
type DType string

//Mode for the entity spcifies certain entity specific characteristics
//Keep this as minimal and add a sub-type for data types such as decimal,boolean,time & date
const (
	TypeString    DType = "S"
	TypeNumber          = "N"
	TypeDateTime        = "T"
	TypeList            = "L"
	TypeReference       = "R"
)

//Dom defines the visual representation of the field
type Dom string

//const defines the types of visual representation dom
const (
	DomText          Dom = "TE"
	DomTextArea          = "TA"
	DomStatus            = "ST"
	DomAutoSelect        = "AS" // same as select but with the twist for auto fill. refer status
	DomAutoComplete      = "AC"
	DomSelect            = "SE" // the default dom for the reference field units. This type mandates the choices limit to 20
	DomMultiSelect       = "MS"
	DomDate              = "DA"
	DomTime              = "TI"
	DomMinute            = "MI"
	DomReminder          = "RE"
	DomDueBy             = "DB"
	DomContact           = "CO" //type which as avatar,name,email and id
	DomOwnCon            = "OC" //type which as avatar,name,email and id + two entities clubed
	DomNotApplicable     = "NA" // the dom for the reference field with no UI needed
	DomImage             = "IM"
)

const (
	PipeReferenceID     = "pipelines"
	FlowsReferenceID    = "flows"
	UsersReferenceID    = "users"
	AccountsReferenceID = "accounts"
)

//field_unit expression
const (
	FuExpNone = "none"
	FuExpDone = "done"
	FuExpPos  = "pos" //set this on positive expression of due_by
	FuExpNeg  = "neg" //set this on negative expression of the due_by
)

const (
	RefTypeAbsolute = ""         //respective childreans will be visible from src/dst details page (from contacts's detail - view deals associated & vice-versa)
	RefTypeStraight = "STRAIGHT" //only the src entity childrean will be visible (from deal's detail - view contacts associated)
	RefTypeReverse  = "REVERSE"  //only the dst entity childrean will be visible (from contacts's detail - view status/owner associated)
)

const (
	MetaKeyDisplayGex  = "display_gex"
	MetaKeyEmailGex    = "email_gex"
	MetaKeyAvatarGex   = "avatar_gex"
	MetaKeyHidden      = "hidden"
	MetaKeyLayout      = "layout"
	MetaKeyFlow        = "flow"
	MetaKeyNode        = "node"
	MetaKeyConfig      = "config"
	MetaKeyLoadChoices = "load_choices"
	MetaKeyRow         = "row"
	MetaMultiChoice    = "multi"
	MetaKeyHTML        = "html"
)

const (
	MetaLayoutTitle    = "title"
	MetaLayoutSubTitle = "sub-title"
	MetaLayoutUsers    = "users"
	MetaLayoutDate     = "date"
)

const (
	WhoStatus    = "status"
	WhoReminder  = "reminder"
	WhoDueBy     = "dueby"
	WhoCloseDate = "close_date"
	WhoAssignee  = "assignee"
	WhoOwner     = "owner"
	WhoFollower  = "follower"
	WhoAvatar    = "avatar"
)

// Field represents structural format of attributes in entity
type Field struct {
	Name         string            `json:"name" validate:"required"`
	DisplayName  string            `json:"display_name" validate:"required"`
	Key          string            `json:"key" validate:"required"`
	Value        interface{}       `json:"value" validate:"required"`
	DataType     DType             `json:"data_type" validate:"required"`
	DomType      Dom               `json:"dom_type" validate:"required"`
	Field        *Field            `json:"field"`
	Meta         map[string]string `json:"meta"` //shall we move the extra prop to this meta or shall we keep it flat?
	Choices      []Choice          `json:"choices"`
	RefID        string            `json:"ref_id"` // this could be another entity_id for reference, pipeline_id for odd with pipleline/playbook
	RefType      string            `json:"ref_type"`
	Dependent    *Dependent        `json:"dependent"`     // if exists, then the results of this field should be filtered by the value of the parent_key specified over the reference_key on the refID specified
	Who          string            `json:"who"`           // who specifies the exact field function. such as:
	UnlinkOffset int               `json:"unlink_offset"` // useful for the graphDB
}

type FieldMeta struct {
	Unique      string `json:"unique"`
	Mandatory   string `json:"mandatory"`
	Hidden      string `json:"hidden"`
	Config      string `json:"config"`     //UI property useful only during display
	Expression  string `json:"expression"` //expression is a double purpose property - executes the checks like, field.value > 100 < 200 or field.value == 'vijay' during "save", checks the operator during segmenting
	Link        string `json:"link"`       //useful for autocomplete. If number of choices greater than 100
	DisplayGex  string `json:"display_gex"`
	EmailGex    string `json:"email_gex"`
	Layout      string `json:"layout"`
	Flow        string `json:"flow"`
	Node        string `json:"node"`
	LoadChoices string `json:"load_choices"`
	Row         string `json:"row"`
}

type Choice struct {
	ID           string      `json:"id"`
	ParentIDs    []string    `json:"parent_ids"`
	Value        interface{} `json:"value"`
	DisplayValue interface{} `json:"display_value"`
	BaseChoice   bool        `json:"base_choice"`
	Default      bool        `json:"default"`
	Verb         string      `json:"verb"` // are we still using this??
}

type Dependent struct {
	ParentKey   string   `json:"parent_key"`
	Expressions []string `json:"expressions"` // if expression exist, execute it to know postive/negative
	Actions     []string `json:"actions"`     // execute the action based on the expression result
}

func (e Entity) FieldsIgnoreError() []Field {
	fields, err := e.Fields()
	if err != nil {
		log.Println(err)
	}
	return fields
}

func (e Entity) ValueAdd(itemFields map[string]interface{}) []Field {
	entityFields := e.FieldsIgnoreError()
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
		return make([]Field, 0), errors.Wrapf(err, "error while unmarshalling entity attributes to fields type %q", e.ID)
	}
	return fields, nil
}

func (e Entity) FilteredFields() ([]Field, error) {
	tmp := make([]Field, 0)
	fields, err := unmarshalFields(e.Fieldsb)
	if err != nil {
		return make([]Field, 0), errors.Wrapf(err, "error while unmarshalling entity attributes to fields type %q", e.ID)
	}

	for _, f := range fields {
		if !f.IsHidden() {
			tmp = append(tmp, f)
		}
	}

	return tmp, nil
}

func (e Entity) WhoFields() map[string]string {
	tmp := make(map[string]string, 0)
	fields, err := unmarshalFields(e.Fieldsb)
	if err != nil {
		return tmp
	}

	for _, f := range fields {
		tmp[f.Who] = f.Key
	}

	return tmp
}

func (e Entity) Key(name string) string {
	fields, err := e.Fields()
	if err != nil {
		return ""
	}
	return NamedKeysMap(fields)[name]
}

func (e Entity) FlowField() *Field {
	fields, _ := e.Fields()
	for _, f := range fields {
		if f.IsFlow() {
			return &f
		}
	}
	return nil
}

func (e Entity) NodeField() *Field {
	fields, _ := e.Fields()
	for _, f := range fields {
		if f.IsNode() {
			return &f
		}
	}
	return nil
}

func (f *Field) SetDisplayGex(key string) {
	if f.Meta == nil {
		f.Meta = make(map[string]string, 0)
	}
	f.Meta[MetaKeyDisplayGex] = key
}

func (f *Field) SetEmailGex(key string) {
	if f.Meta == nil {
		f.Meta = make(map[string]string, 0)
	}
	f.Meta[MetaKeyEmailGex] = key
}

func (f Field) isConfig() bool {
	if val, ok := f.Meta[MetaKeyConfig]; ok && val == "true" {
		return true
	}
	return false
}

func (f Field) IsFlow() bool {
	if val, ok := f.Meta[MetaKeyFlow]; ok && val == "true" {
		return true
	}
	return false
}

func (f Field) IsNode() bool {
	if val, ok := f.Meta[MetaKeyNode]; ok && val == "true" {
		return true
	}
	return false
}

func (f Field) ForceLoadChoices() bool {
	if val, ok := f.Meta[MetaKeyLoadChoices]; ok && val == "true" {
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

func (f Field) IsDateTime() bool {
	if f.DataType == TypeDateTime {
		return true
	}
	return false
}

func (f Field) IsList() bool {
	if f.DataType == TypeList {
		return true
	}
	return false
}

func (f Field) IsDependent() bool {
	if f.Dependent != nil {
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

func (f Field) SetMeta(key string) {
	f.Meta[key] = "true"
}

func (f Field) ValidRefField() bool {
	return f.IsReference() && f.Value != nil && f.Value != "" && len(f.Value.([]interface{})) > 0
}

func (f Field) RefValues() []interface{} {
	return f.Value.([]interface{})
}

func (f Field) DisplayGex() string {
	if val, ok := f.Meta[MetaKeyDisplayGex]; ok {
		return val
	}
	return ""
}

func (f Field) EmailGex() string {
	if val, ok := f.Meta[MetaKeyEmailGex]; ok {
		return val
	}
	return ""
}

func (f Field) IsTitleLayout() bool {
	if val, ok := f.Meta[MetaKeyLayout]; ok {
		return val == "title"
	}
	return false
}

func (f Field) IsHidden() bool {
	if val, ok := f.Meta[MetaKeyHidden]; ok {
		return val == "true"
	}
	return false
}

func FieldsMap(entityFields []Field) map[string]interface{} {
	params := map[string]interface{}{}
	for _, f := range entityFields {
		params[f.Key] = f.Value
	}
	return params
}

func NamedKeysMap(entityFields []Field) map[string]string {
	params := map[string]string{}
	for _, f := range entityFields {
		params[f.Name] = f.Key
	}
	return params
}

func NamedFieldsObjMap(entityFields []Field) map[string]Field {
	params := map[string]Field{}
	for _, f := range entityFields {
		params[f.Name] = f
	}
	return params
}

func MetaFieldsObjMap(entityFields []Field) map[string]Field {
	params := map[string]Field{}
	for _, f := range entityFields {
		params[f.Meta[MetaKeyLayout]] = f
	}
	return params
}

func KeyedFieldsObjMap(entityFields []Field) map[string]Field {
	params := map[string]Field{}
	if entityFields != nil { // does this check needed
		for _, f := range entityFields {
			params[f.Key] = f
		}
	}
	return params
}

func (f Field) ChoicesValues() []string {
	s := make([]string, len(f.Choices))
	for i, choice := range f.Choices {
		s[i] = fmt.Sprint(choice.Value)
	}
	return s
}

func unmarshalFields(fieldsB string) ([]Field, error) {
	var fields []Field
	if err := json.Unmarshal([]byte(fieldsB), &fields); err != nil {
		return nil, err
	}
	return fields, nil
}

func refFields(fields []Field) map[string]relationship.Relatable {
	referenceFieldsMap := make(map[string]relationship.Relatable, 0)
	for _, f := range fields {
		if f.IsReference() { // TODO: also check if customer explicitly asks for it. Don't do this for all the reference fields
			if f.RefType == RefTypeStraight {
				referenceFieldsMap[f.Key] = relationship.MakeRelatable(f.RefID, relationship.RTypeStraight)
			} else if f.RefType == RefTypeReverse {
				referenceFieldsMap[f.Key] = relationship.MakeRelatable(f.RefID, relationship.RTypeReverse)
			} else if f.RefType == RefTypeAbsolute {
				referenceFieldsMap[f.Key] = relationship.MakeRelatable(f.RefID, relationship.RTypeAbsolute)
			}
		}
	}
	return referenceFieldsMap
}
