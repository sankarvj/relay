package draft

import (
	"time"

	"github.com/lib/pq"
)

const (
	TeamCSP = "csp"
	TeamCRP = "crp"
	TeamEMP = "emp"
)

var TeamDomainMap = map[string]string{
	TeamCSP: "csp.workbaseone.com",
	TeamCRP: "crp.workbaseone.com",
	TeamEMP: "emp.workbaseone.com",
}

type Draft struct {
	ID            string         `db:"draft_id" json:"id"`
	AccountName   string         `db:"account_name" json:"account_name"`
	BusinessEmail string         `db:"business_email" json:"business_email"`
	Host          string         `db:"host" json:"host"`
	Teams         pq.StringArray `db:"teams" json:"teams"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt     int64          `db:"updated_at" json:"updated_at"`
}

type NewDraft struct {
	AccountName   string   `json:"account_name" validate:"required"`
	BusinessEmail string   `json:"business_email" validate:"required"`
	Teams         []string `json:"teams" validate:"required"`
	Host          string   `json:"host" validate:"required"`
}
