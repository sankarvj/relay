package draft

import (
	"time"

	"github.com/lib/pq"
)

const (
	TeamCSM = "csm"
	TeamCRM = "crm"
	TeamCSD = "csd"
)

type Draft struct {
	ID            string         `db:"draft_id" json:"id"`
	AccountName   string         `db:"account_name" json:"account_name"`
	BusinessEmail string         `db:"business_email" json:"business_email"`
	Teams         pq.StringArray `db:"teams" json:"teams"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt     int64          `db:"updated_at" json:"updated_at"`
}

type NewDraft struct {
	AccountName   string   `json:"account_name" validate:"required"`
	BusinessEmail string   `json:"business_email" validate:"required"`
	Teams         []string `json:"teams" validate:"required"`
}
