package conversation

import (
	"encoding/json"
	"log"
)

type Message struct {
	Action   string                `json:"action"`
	Payload  ViewModelConversation `json:"payload"`
	Room     string                `json:"room"`
	User     string                `json:"user"`
	ClientID string                `json:"client_id"`
}

type ViewModelConversation struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	UserName   string `json:"user_name"`
	UserAvatar string `json:"user_avatar"`
	Type       int    `json:"type"`
	Message    string `json:"message"`
}

func NewMessage(action string, viewModelConversation ViewModelConversation, room, user, clientID string) *Message {
	return &Message{
		Action:   action,
		Payload:  viewModelConversation,
		Room:     room,
		User:     user,
		ClientID: clientID,
	}
}

func (message *Message) encode() []byte {
	json, err := json.Marshal(message)
	if err != nil {
		log.Println("***> unexpected/unhandled error in internal.platform.conversation. when marshaling message. error:", err)
	}

	return json
}

func (vmc *ViewModelConversation) encode() []byte {
	json, err := json.Marshal(vmc)
	if err != nil {
		log.Println("***> unexpected/unhandled error in internal.platform.conversation. when unmarshaling message. error:", err)
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
