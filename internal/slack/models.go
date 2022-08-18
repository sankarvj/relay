package slack

type Slack struct {
	BotToken string
	Payload
}

type Payload struct {
	Token    string `json:"token"`
	TeamID   string `json:"team_id"`
	APIAppID string `json:"api_app_id"`
	Event    struct {
		Type    string `json:"type"`
		User    string `json:"user"`
		Text    string `json:"text"`
		Ts      string `json:"ts"`
		Channel string `json:"channel"`
		EventTs string `json:"event_ts"`
	} `json:"event"`
	Type        string   `json:"type"`
	EventID     string   `json:"event_id"`
	EventTime   int64    `json:"event_time"`
	AuthedUsers []string `json:"authed_users"`
	Challenge   string   `json:"challenge"`
}

type SlackViewResponse struct {
	Ok      bool   `json:"ok"`
	Error   string `json:"error"`
	Warning string `json:"warning"`
}
