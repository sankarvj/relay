package event

type NewEvent struct {
	Block       string   `json:"block"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Icon        string   `json:"icon"`
}
