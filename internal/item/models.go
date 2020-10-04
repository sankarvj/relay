package item

import (
	"time"
)

// Item represents the individual unit of entity
type Item struct {
	ID        string    `db:"item_id" json:"id"`
	AccountID string    `db:"account_id" json:"account_id"`
	EntityID  string    `db:"entity_id" json:"entity_id"`
	State     int       `db:"state" json:"state"`
	Fieldsb   string    `db:"fieldsb" json:"fieldsb"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt int64     `db:"updated_at" json:"updated_at"`
}

// TimeSeriesItem represents the individual unit of entity
type TimeSeriesItem struct {
	State string    `db:"status" json:"status"`
	Date  time.Time `db:"date" json:"date"`
	Value int64     `db:"value" json:"value"`
}

// ViewModelItem represents the view model of item
// (i.e) it has fields instead of attributes
type ViewModelItem struct {
	ID     string                 `json:"id"`
	Fields map[string]interface{} `json:"fields"`
}

// NewItem has information needed to creat new item
type NewItem struct {
	ID        string                 `json:"id"`
	AccountID string                 `json:"account_id"`
	EntityID  string                 `json:"entity_id"`
	Fields    map[string]interface{} `json:"fields" validate:"required"`
}

// UpdateItem defines what information may be provided to modify an existing
// Item. All fields are optional so clients can send just the fields they want
// changed. It uses pointer fields so we can differentiate between a field that
// was not provided and a field that was provided as explicitly blank. Normally
// we do not want to use pointers to basic types but we make exceptions around
// marshalling/unmarshalling.
type UpdateItem struct {
	Fieldsb *string `json:"fieldsb"`
}
