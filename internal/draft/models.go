package draft

import (
	"time"

	"github.com/lib/pq"
	"gitlab.com/vjsideprojects/relay/internal/team"
)

var TeamDomainMap = map[string]string{
	team.PredefinedTeamCSP: "csp.workbaseone.com",
	team.PredefinedTeamCRP: "crp.workbaseone.com",
	team.PredefinedTeamEMP: "emp.workbaseone.com",
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
	Host          string   `json:"host"`
}
