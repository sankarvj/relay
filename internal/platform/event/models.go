package event

import (
	"encoding/json"
	"log"
)

type ViewModelMessage struct {
	Payload string `json:"payload"`
}

type Message struct {
	Action   string `json:"action"`
	Payload  string `json:"payload"`
	Room     string `json:"room"`
	User     string `json:"user"`
	ClientID string `json:"client_id"`
}

func NewMessage(action, payload, room, user, clientID string) *Message {
	return &Message{
		Action:   action,
		Payload:  payload,
		Room:     room,
		User:     user,
		ClientID: clientID,
	}
}

func (message *Message) encode() []byte {
	json, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}

	return json
}

func (message *Message) UnmarshalJSON(data []byte) error {
	type Alias Message
	msg := &struct {
		Sender Client `json:"sender"`
		*Alias
	}{
		Alias: (*Alias)(message),
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}
	return nil
}

func (vmMessage *ViewModelMessage) encode() []byte {
	json, err := json.Marshal(vmMessage)
	if err != nil {
		log.Println(err)
	}
	return json
}
