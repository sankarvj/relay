package account

import (
	"time"
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
	ID     string `json:"id" validate:"required"`
	Name   string `json:"name" validate:"required"`
	Domain string `json:"domain" validate:"required"`
}
