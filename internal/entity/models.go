package entity

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
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
	Name      string  `json:"name" validate:"required"`
	AccountID string  `json:"account_id" validate:"required"`
	TeamID    string  `json:"team_id" validate:"required"`
	Fields    []Field `json:"fields" validate:"required"`
	Category  int     `json:"category" validate:"required"`
	State     int     `json:"state" validate:"required"`
}

// Field represents structural format of attributes in entity
type Field struct {
	Name        string      `json:"name" validate:"required"`
	DisplayName string      `json:"display_name" validate:"required"`
	Key         string      `json:"key" validate:"required"`
	Value       interface{} `json:"value" validate:"required"`
	DataType    FieldType   `json:"data_type" validate:"required"`
	Unique      bool        `json:"unique"`
	Mandatory   bool        `json:"mandatory"`
	Hidden      bool        `json:"hidden"`
	Config      bool        `json:"config"`
	Reference   string      `json:"reference"`
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

//FieldType defines the type of field
type FieldType string

//Mode for the entity spcifies certain entity specific characteristics
const (
	TypeString       FieldType = "S"
	TypeNumber       FieldType = "N"
	TypeDataTime     FieldType = "DT"
	TypeStatus       FieldType = "ST"
	TypeAutoComplete FieldType = "AC"
	TypeTime         FieldType = "T"
	TypeMinute       FieldType = "M"
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
)

// Fields parses attribures to fields
func (e Entity) Fields() ([]Field, error) {
	fields, err := e.AllFields()
	if err != nil {
		return nil, err
	}
	//remove all config fields
	temp := fields[:0]
	for _, field := range fields {
		if !field.Config {
			temp = append(temp, field)
		}
	}
	fields = temp
	return fields, nil
}

// AllFields parses attribures to fields
func (e Entity) AllFields() ([]Field, error) {
	var fields []Field
	if err := json.Unmarshal([]byte(e.Fieldsb), &fields); err != nil {
		return nil, errors.Wrapf(err, "error while unmarshalling entity attributes to fields type %q", e.ID)
	}
	return fields, nil
}
