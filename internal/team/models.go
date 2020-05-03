package team

import "time"

// Team represents sub domains of an organisation
type Team struct {
	ID          int64     `db:"team_id" json:"id"`
	AccountID   string    `db:"account_id" json:"account_id"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   int64     `db:"updated_at" json:"updated_at"`
}

// NewTeam contains information needed to create a new Team.
type NewTeam struct {
	AccountID string `json:"account_id"`
	Name      string `json:"name" validate:"required"`
}
