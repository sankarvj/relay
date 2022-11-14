package account

import (
	"time"
)

const (
	PlanFree    = 0
	PlanStartup = 1
	PlanPro     = 2
)

// Account represents the organization where set of users belong
type Account struct {
	ID              string    `db:"account_id" json:"id"`
	ParentAccountID *string   `db:"parent_account_id" json:"parent_account_id"`
	Name            string    `db:"name" json:"name"`
	Domain          string    `db:"domain" json:"domain"`
	Avatar          *string   `db:"avatar" json:"avatar"`
	Plan            int       `db:"plan" json:"plan"`
	Mode            int       `db:"mode" json:"mode"`
	CustomerMail    string    `db:"cus_mail" json:"cus_mail"`
	CustomerID      string    `db:"cus_id" json:"cus_id"`
	TimeZone        *string   `db:"timezone" json:"timezone"`
	Language        *string   `db:"language" json:"language"`
	Country         *string   `db:"country" json:"country"`
	IssuedAt        time.Time `db:"issued_at" json:"issued_at"`
	Expiry          time.Time `db:"expiry" json:"expiry"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       int64     `db:"updated_at" json:"updated_at"`
}

// NewAccount contains information needed to create a new Account.
type NewAccount struct {
	ID      string `json:"id" validate:"required"`
	Name    string `json:"name" validate:"required"`
	Domain  string `json:"domain"`
	DraftID string `json:"draft_id"`
}

type LaunchAccount struct {
	DraftID           string `json:"draft_id" validate:"required"`
	BusinessEmailHash string `json:"business_email_hash" validate:"required"`
	FirstName         string `json:"first_name" validate:"required"`
	LastName          string `json:"last_name"`
}
