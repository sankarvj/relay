package entity

import (
	"time"

	"github.com/lib/pq"
)

// Entity represents the building block of all the tasks
type Entity struct {
	ID          string         `db:"entity_id" json:"id"`
	AccountID   string         `db:"account_id" json:"account_id"`
	TeamID      string         `db:"team_id" json:"team_id"`
	Name        string         `db:"name" json:"name"`
	Description *string        `db:"description" json:"description"`
	Category    int            `db:"category" json:"category"`
	State       int            `db:"state" json:"state"`
	Status      int            `db:"status" json:"status"`
	Retry       int            `db:"retry" json:"retry"`
	Fieldsb     string         `db:"fieldsb" json:"fieldsb"`
	Tags        pq.StringArray `db:"tags" json:"tags"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   int64          `db:"updated_at" json:"updated_at"`
}

// ViewModelEntity represents the view model of entity
// (i.e) it has fields instead of attributes
type ViewModelEntity struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Category    int       `json:"category"`
	State       int       `json:"state"`
	Status      int       `json:"status"`
	Fields      []Field   `json:"fields"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   int64     `json:"updated_at"`
}

// NewEntity has information needed to creat new entity
type NewEntity struct {
	ID        string  `json:"id"`
	Name      string  `json:"name" validate:"required"`
	AccountID string  `json:"account_id"`
	TeamID    string  `json:"team_id"`
	Fields    []Field `json:"fields" validate:"required"`
	Category  int     `json:"category"`
	State     int     `json:"state"`
}

// EmailEntity represents structural format of email entity
type EmailEntity struct {
	Domain  string `json:"domain"`
	APIKey  string `json:"api_key"`
	Sender  string `json:"sender"`
	To      string `json:"to"`
	Cc      string `json:"cc"`
	Bcc     string `json:"bcc"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

//DelayEntity represents the structural format of delay entity
type DelayEntity struct {
	DelayBy string `json:"delay_by"`
	Repeat  string `json:"repeat"`
}

// WebHookEntity represents structural format of webhook entity
type WebHookEntity struct {
	Path    string            `json:"path"`
	Host    string            `json:"host"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
}

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
	DomAutoSelect        = "AS" // same as select but with the twist for auto fill.
	DomDate              = "DA"
	DomTime              = "TI"
	DomMinute            = "MI"
	DomMultiSelect       = "MS"
	DomNotApplicable     = "NA" // the dom for the reference field with no UI needed
	DomPipeline          = "PL"
	DomPlayBook          = "PB"
	DomReminder          = "RE"
)

//State for the entity specifies the current state of the entity
const (
	StateUnknown = 0
)

//Category specifies the different type of entities
const (
	CategoryUnknown    = 0
	CategoryData       = 1
	CategoryAPI        = 2
	CategoryTimeSeries = 3
	CategoryEmail      = 4
	CategoryUserSeries = 5
	CategorySchedule   = 6
	CategoryDelay      = 7
	CategoryChildUnit  = 8
	CategoryTask       = 9
)

//field_unit expression
const (
	FuExpNone   = "none"
	FuExpPos    = "pos"    //set this on positive expression of due_by
	FuExpNeg    = "neg"    //set this on negative expression of the due_by
	FuExpManual = "manual" //keep as it is unless manually changes
)
