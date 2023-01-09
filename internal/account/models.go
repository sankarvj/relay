package account

import (
	"time"
)

const (
	PlanFree    = 0
	PlanStartup = 1
	PlanPro     = 2
)

const (
	StatusTrial  = "trialing"
	StatusActive = "active"
)

// Account represents the organization where set of users belong
type Account struct {
	ID              string    `db:"account_id" json:"id"`
	ParentAccountID *string   `db:"parent_account_id" json:"parent_account_id"`
	Name            string    `db:"name" json:"name"`
	Domain          string    `db:"domain" json:"domain"`
	Avatar          *string   `db:"avatar" json:"avatar"`
	Mode            int       `db:"mode" json:"mode"`
	CustomerID      *string   `db:"cus_id" json:"cus_id"`
	CustomerMail    *string   `db:"cus_mail" json:"cus_mail"`
	CustomerSeat    int       `db:"cus_seat" json:"cus_seat"`
	CustomerStatus  string    `db:"cus_status" json:"cus_status"`
	CustomerPlan    int       `db:"cus_plan" json:"cus_plan"`
	TrailStart      float64   `db:"trail_start" json:"trail_start"`
	TrailEnd        float64   `db:"trail_end" json:"trail_end"`
	TimeZone        *string   `db:"timezone" json:"timezone"`
	Language        *string   `db:"language" json:"language"`
	Country         *string   `db:"country" json:"country"`
	UseDB           string    `db:"use_db" json:"use_db"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       int64     `db:"updated_at" json:"updated_at"`
}

// NewAccount contains information needed to create a new Account.
type NewAccount struct {
	ID             string  `json:"id" validate:"required"`
	Name           string  `json:"name" validate:"required"`
	Domain         string  `json:"domain"`
	DraftID        string  `json:"draft_id"`
	CustomerPlan   int     `json:"cus_plan"`
	CustomerStatus string  `json:"cus_status"`
	TrailStart     float64 `json:"trail_start"`
	TrailEnd       float64 `json:"trail_end"`
	UseDB          string  `json:"use_db"`
}

type LaunchAccount struct {
	DraftID           string `json:"draft_id" validate:"required"`
	BusinessEmailHash string `json:"business_email_hash" validate:"required"`
	FirstName         string `json:"first_name" validate:"required"`
	LastName          string `json:"last_name"`
}
