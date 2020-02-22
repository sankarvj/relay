package entity

import (
	"time"

	"github.com/lib/pq"
)

// Entity represents the building block of all the tasks
type Entity struct {
	ID          string         `db:"entity_id" json:"id"`
	TeamID      int64          `db:"team_id" json:"team_id"`
	Name        string         `db:"name" json:"name"`
	Description *string        `db:"description" json:"description"`
	State       int            `db:"state" json:"state"`
	Mode        int            `db:"mode" json:"mode"`
	Retry       int            `db:"retry" json:"retry"`
	Attributes  string         `db:"attributes" json:"attributes"`
	Tags        pq.StringArray `db:"tags" json:"tags"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   int64          `db:"updated_at" json:"updated_at"`
}

// ViewModelEntity represents the view model of entity
// (i.e) it has fields instead of attributes
type ViewModelEntity struct {
	ID          string    `json:"id"`
	TeamID      int64     `json:"team_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	State       int       `json:"state"`
	Mode        int       `json:"mode"`
	Retry       int       `json:"retry"`
	Fields      []Field   `json:"fields"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   int64     `json:"updated_at"`
}

// NewEntity has information needed to creat new entity
type NewEntity struct {
	TeamID string  `json:"team_id"`
	Name   string  `json:"name" validate:"required"`
	Fields []Field `json:"fields" validate:"required"`
}

// Field represents structural format of attributes in entity
type Field struct {
	Name     string `json:"name" validate:"required"`
	Key      string `json:"key" validate:"required"`
	Value    string `json:"value" validate:"required"`
	DataType string `json:"data_type" validate:"required"`
}

//Mode for the entity spcifies certain entity specific characteristics
const (
	ModePrimary = 1
)
