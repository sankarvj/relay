package event

type NewEvent struct {
	Block      string                 `json:"block"`
	Identifier string                 `json:"identifier"`
	Properties map[string]interface{} `json:"properties"`
}
