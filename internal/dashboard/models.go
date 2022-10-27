package dashboard

import "time"

type Type string

const (
	TypeDefault Type = "default"
	TypeHome    Type = "home"
)

type Dashboard struct {
	ID        string    `db:"dashboard_id" json:"id"`
	AccountID string    `db:"account_id" json:"account_id"`
	TeamID    string    `db:"team_id" json:"team_id"`
	EntityID  string    `db:"entity_id" json:"entity_id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Name      string    `db:"name" json:"name"`
	Type      string    `db:"type" json:"type"`
	Metab     string    `db:"metab" json:"metab"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type NewDashboard struct {
	ID        string            `json:"dashboard_id"`
	AccountID string            `json:"account_id"`
	TeamID    string            `json:"team_id"`
	EntityID  string            `json:"entity_id"`
	UserID    string            `json:"user_id"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Meta      map[string]string `json:"meta"`
}
