package item

import "time"

// Item represents the individual unit of entity
type Item struct {
	ID           string    `db:"item_id" json:"id"`
	ParentItemID *string   `db:"parent_item_id" json:"parent_item_id"`
	EntityID     string    `db:"entity_id" json:"entity_id"`
	State        int       `db:"state" json:"state"`
	Input        string    `db:"input" json:"input"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    int64     `db:"updated_at" json:"updated_at"`
}

// ViewModelItem represents the view model of item
// (i.e) it has fields instead of attributes
type ViewModelItem struct {
	ID     string  `json:"id"`
	Fields []Field `json:"fields"`
}

// NewItem has information needed to creat new item
type NewItem struct {
	Fields []Field `json:"fields" validate:"required"`
}

// Field represents structural format of attributes in entity
type Field struct {
	Name     string `json:"name" validate:"required"`
	Key      string `json:"key" validate:"required"`
	Value    string `json:"value" validate:"required"`
	DataType string `json:"data_type" validate:"required"`
}

// UpdateItem defines what information may be provided to modify an existing
// Item. All fields are optional so clients can send just the fields they want
// changed. It uses pointer fields so we can differentiate between a field that
// was not provided and a field that was provided as explicitly blank. Normally
// we do not want to use pointers to basic types but we make exceptions around
// marshalling/unmarshalling.
type UpdateItem struct {
	Input *string `json:"input"`
}
