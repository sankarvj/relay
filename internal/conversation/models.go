package conversation

import "time"

const (
	TypeEmailSent     = 0 //actual email - first message sent
	TypeEmailReceived = 1 //actual email - first message received
	TypeConvSent      = 2 //subsequent message sent for the first message
	TypeConvReceived  = 3 //subsequent message received for the first message
)

const (
	StateNotSent   = 0
	StateSent      = 1
	StateDelivered = 2
	StateOpened    = 3
)

type Conversation struct {
	ID        string    `db:"conversation_id" json:"id"`
	AccountID string    `db:"account_id" json:"account_id"`
	EntityID  string    `db:"entity_id" json:"entity_id"`
	ItemID    *string   `db:"item_id" json:"item_id"` // * because it could be null
	UserID    string    `db:"user_id" json:"user_id"`
	Type      int       `db:"type" json:"type"`
	State     int       `db:"state" json:"state"`
	Message   string    `db:"message" json:"message"`
	Payload   string    `db:"payload" json:"payload"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt int64     `db:"updated_at" json:"updated_at"`
}

type NewConversation struct {
	ID        string                 `json:"id"`
	AccountID string                 `json:"account_id"`
	EntityID  string                 `json:"entity_id"`
	ItemID    *string                `json:"item_id"`
	UserID    string                 `json:"user_id"`
	Type      int                    `json:"type"`
	State     int                    `json:"state"`
	Message   string                 `json:"message" validate:"required"`
	Payload   map[string]interface{} `json:"payload"`
}

type ConversationUsr struct {
	ID         string    `db:"conversation_id" json:"id"`
	UserID     *string   `db:"user_id" json:"user_id"`
	UserName   *string   `db:"user_name" json:"user_name"`
	UserAvatar *string   `db:"user_avatar" json:"user_avatar"`
	Type       int       `db:"type" json:"type"`
	State      int       `db:"state" json:"state"`
	Message    string    `db:"message" json:"message"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type ViewModelConversation struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	UserName   string    `json:"user_name"`
	UserAvatar string    `json:"user_avatar"`
	Type       int       `json:"type"`
	State      int       `json:"state"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}
