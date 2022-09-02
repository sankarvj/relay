package team

import "time"

// Team represents sub domains of an organisation
type Team struct {
	ID          string    `db:"team_id" json:"id"`
	AccountID   string    `db:"account_id" json:"account_id"`
	Name        string    `db:"name" json:"name"`               //use this internally for indentifying the team template CRM,CSM,EM,PM
	Description *string   `db:"description" json:"description"` //use this for UI display
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   int64     `db:"updated_at" json:"updated_at"`
}

// NewTeam contains information needed to create a new Team.
type NewTeam struct {
	AccountID string   `json:"account_id"`
	Name      string   `json:"name" validate:"required"`
	Modules   []string `json:"modules"`
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
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var templatesMap = map[string]string{
	"crm":             "CRM",
	"support":         "Support Desk",
	"onboarding-emp":  "Employee Onboarding",
	"onboarding-cust": "Customer Onboarding",
	"project":         "Project Management",
}

var templatesDescMap = map[string]string{
	"crm":             "Customer relationship management",
	"support":         "Customer support desk",
	"onboarding-emp":  "Onboard your employees with sequence of steps",
	"onboarding-cust": "Onboard your customers with sequence of steps",
	"project":         "Manage your project tasks",
}
