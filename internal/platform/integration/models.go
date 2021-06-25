package integration

type Meeting struct {
	ID          string   `json:"id"`
	CalID       string   `json:"cal_id"`
	Summary     string   `json:"summary,omitempty"`
	HtmlLink    string   `json:"htmlLink,omitempty"`
	Description string   `json:"description,omitempty"`
	Attendees   []string `json:"attendees"`
	StartTime   string   `json:"start_time,omitempty"`
	EndTime     string   `json:"end_time,omitempty"`
	TimeZone    string   `json:"timeZone,omitempty"`
	Created     string   `json:"created,omitempty"`
	Updated     string   `json:"updated,omitempty"`
}
