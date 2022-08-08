package visitor

import "time"

const (
	DefaultUserID = "00000000-0000-0000-0000-000000000000"
)

type Visitor struct {
	VistitorID string    `db:"visitor_id" json:"visitor_id"`
	AccountID  string    `db:"account_id" json:"account_id"`
	TeamID     string    `db:"team_id" json:"team_id"`
	EntityID   string    `db:"entity_id" json:"entity_id"`
	ItemID     string    `db:"item_id" json:"item_id"`
	Name       string    `db:"name" json:"name"`
	Email      string    `db:"email" json:"email"`
	Token      string    `db:"token" json:"token"`
	Active     bool      `db:"active" json:"active"`
	SignedIn   bool      `db:"signed_in" json:"signed_in"`
	ExpireAt   time.Time `db:"expire_at" json:"expire_at"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  int64     `db:"updated_at" json:"updated_at"`
}

type NewVisitor struct {
	AccountID string `json:"account_id"`
	TeamID    string `json:"team_id"`
	EntityID  string `json:"entity_id"`
	ItemID    string `json:"item_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Token     string `json:"token"`
}
