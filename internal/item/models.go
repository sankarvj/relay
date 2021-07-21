package item

import (
	"time"
)

// Item represents the individual unit of entity
type Item struct {
	ID        string    `db:"item_id" json:"id"`
	AccountID string    `db:"account_id" json:"account_id"`
	EntityID  string    `db:"entity_id" json:"entity_id"`
	GenieID   *string   `db:"genie_id" json:"genie_id"` // * because it could be null
	UserID    *string   `db:"user_id" json:"user_id"`   // * because it could be null
	State     int       `db:"state" json:"state"`
	Type      int       `db:"type" json:"type"`
	Name      *string   `db:"name" json:"name"`
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

// NewItem has information needed to creat new item
type NewItem struct {
	ID        string                 `json:"id"`
	AccountID string                 `json:"account_id"`
	EntityID  string                 `json:"entity_id"`
	GenieID   *string                `json:"genie_id"`
	UserID    *string                `json:"user_id"`
	Name      *string                `json:"name"`
	Type      int                    `json:"type"`
	State     int                    `json:"state"`
	Fields    map[string]interface{} `json:"fields" validate:"required"`
	Source    map[string]string      `json:"source"`
}

type RefItem struct {
	ID      string `json:"id"`
	Display string `json:"display"`
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

//State for the item specifies when to associate a item to the lists
const (
	StateDefault   = 0
	StateBluePrint = 1 //used when adding blueprint item in the workflow
)

//Type for the item is still open we can use it for anything
const (
	TypeDefault = 0
)
