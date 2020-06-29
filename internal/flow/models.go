package flow

import "time"

// Flow represents a single workflow
type Flow struct {
	ID          string    `db:"flow_id" json:"id"`
	AccountID   string    `db:"account_id" json:"account_id"`
	EntityID    string    `db:"entity_id" json:"entity_id"`
	Expression  string    `db:"expression" json:"expression"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Type        int       `db:"type" json:"type"`
	Status      int       `db:"status" json:"status"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   int64     `db:"updated_at" json:"updated_at"`
}

// ViewModelFlow represents the view model of flow
type ViewModelFlow struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Expression  string `json:"condition"`
}

// NewFlow has information needed to creat new flow
type NewFlow struct {
	AccountID   string `json:"account_id" validate:"required"`
	EntityID    string `json:"entity_id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Expression  string `json:"expression" validate:"required"`
	Type        int    `json:"type" validate:"required"`
}
