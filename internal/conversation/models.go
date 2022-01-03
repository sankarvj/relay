package conversation

import "time"

type Conversation struct {
	ID        string    `db:"conversation_id" json:"id"`
	AccountID string    `db:"account_id" json:"account_id"`
	EntityID  string    `db:"entity_id" json:"entity_id"`
	ItemID    *string   `db:"item_id" json:"item_id"` // * because it could be null
	UserID    string    `db:"user_id" json:"user_id"`
	Type      int       `db:"type" json:"type"`
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
	Message   string                 `json:"message" validate:"required"`
	Payload   map[string]interface{} `json:"payload"`
}

type ViewModelConversation struct {
	ID         string  `db:"conversation_id" json:"id"`
	UserID     *string `db:"user_id" json:"user_id"`
	UserName   *string `db:"user_name" json:"user_name"`
	UserAvatar *string `db:"user_avatar" json:"user_avatar"`
	Type       int     `db:"type" json:"type"`
	Message    string  `db:"message" json:"message"`
}
