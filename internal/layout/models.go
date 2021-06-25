package layout

import "time"

const (
	DefaultUserID = "00000000-0000-0000-0000-000000000000"
)

type Layout struct {
	Name      string    `db:"name" json:"name"`
	AccountID string    `db:"account_id" json:"account_id"`
	EntityID  string    `db:"entity_id" json:"entity_id"`
	UserID    *string   `db:"user_id" json:"user_id"` // * because it could be null
	Type      int       `db:"type" json:"type"`
	Fieldsb   string    `db:"fieldsb" json:"fieldsb"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt int64     `db:"updated_at" json:"updated_at"`
}

type NewLayout struct {
	Name      string            `json:"name"`
	AccountID string            `json:"account_id"`
	EntityID  string            `json:"entity_id"`
	UserID    *string           `json:"user_id"`
	Type      int               `json:"type"`
	Fields    map[string]string `json:"fields" validate:"required"`
}
