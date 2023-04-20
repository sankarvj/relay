package team

import "time"

const (
	PredefinedTeamCSP  = "csp"
	PredefinedTeamCRP  = "crp"
	PredefinedTeamEMP  = "emp"
	PredefinedTeamPMP  = "pmp"
	PredefinedTeamCSup = "csup"
	PredefinedTeamINC  = "inc"
)

// Team represents sub domains of an organisation
type Team struct {
	ID          string    `db:"team_id" json:"id"`
	AccountID   string    `db:"account_id" json:"account_id"`
	LookUp      string    `db:"look_up" json:"look_up"`         //use this internally for indentifying the team template CRM,CSM,EM,PM
	Name        string    `db:"name" json:"name"`               //use this for UI display
	Description *string   `db:"description" json:"description"` //use this for UI display
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   int64     `db:"updated_at" json:"updated_at"`
}

// NewTeam contains information needed to create a new Team.
type NewTeam struct {
	AccountID   string   `json:"account_id"`
	LookUp      string   `json:"look_up" validate:"required"`
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description"`
	Modules     []string `json:"modules"`
}

type Module struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Checked bool   `json:"checked"`
}

var modulesMap = map[string]string{
	"deals":     "Deals",
	"tasks":     "Tasks",
	"meetings":  "Meetings",
	"notes":     "Notes",
	"items":     "Items",
	"leads":     "Leads",
	"employees": "Employees",
}

type Template struct {
	Icon        string `json:"icon"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var templatesNameMap = map[string]string{
	PredefinedTeamCSP:  "Success",
	PredefinedTeamCRP:  "Sales",
	PredefinedTeamEMP:  "Employee",
	PredefinedTeamPMP:  "Project",
	PredefinedTeamCSup: "Support",
	PredefinedTeamINC:  "Incident",
}

var templatesDescMap = map[string]string{
	PredefinedTeamCSP:  "Customer success platform",
	PredefinedTeamCRP:  "Customer relationship platform",
	PredefinedTeamEMP:  "Employee management platform",
	PredefinedTeamPMP:  "Project management platform",
	PredefinedTeamCSup: "Customer support platform",
	PredefinedTeamINC:  "Incident management platform",
}

var templatesIconMap = map[string]string{
	PredefinedTeamCSP:  "csp.svg",
	PredefinedTeamCRP:  "crp.svg",
	PredefinedTeamEMP:  "emp.svg",
	PredefinedTeamPMP:  "pmp.svg",
	PredefinedTeamCSup: "csup.svg",
	PredefinedTeamINC:  "inc.svg",
}
