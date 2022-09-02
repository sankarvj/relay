package stream

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

//Type for the item is still open we can use it for anything
const (
	TypeDefault                = 0
	TypeItemCreate             = 1
	TypeItemUpdate             = 2
	TypeItemDelete             = 3
	TypeItemRemind             = 4
	TypeItemDelayed            = 5
	TypeEmailConversationAdded = 6
	TypeChatConversationAdded  = 7
)

const (
	StateQueued          = 0
	StateWorkflow        = 1
	StateConnection      = 2
	StateCategory        = 3
	StateWho             = 4
	StateNotification    = 5
	StateRedis           = 6
	StatePrimaryDBDelete = 7
	StateSecDBDelete     = 8
)

type Message struct {
	ID             string                 `json:"id"`
	Type           int                    `json:"type"`
	AccountID      string                 `json:"account_id"`
	UserID         string                 `json:"user_id"`
	EntityID       string                 `json:"entity_id"`
	ItemID         string                 `json:"item_id"`
	ConversationID string                 `json:"conversation_id"`
	NewFields      map[string]interface{} `json:"new_fields"`
	OldFields      map[string]interface{} `json:"old_fields"`
	Source         map[string][]string    `json:"source"`
	Meta           map[string]interface{} `json:"meta"`
	Comment        string                 `json:"comment"`
	State          int                    `json:"state"`
}

func NewCreteItemMessage(ctx context.Context, db *sqlx.DB, accountID, userID, entityID, itemID string, source map[string][]string) *Message {
	m := &Message{
		ID:        fmt.Sprintf("%s::->>%s", "create", uuid.New().String()),
		Type:      TypeItemCreate,
		AccountID: accountID,
		UserID:    userID,
		EntityID:  entityID,
		ItemID:    itemID,
		Source:    source,
		State:     StateQueued,
	}
	add(ctx, db, m, "Queued", StateQueued)
	return m
}

func NewUpdateItemMessage(ctx context.Context, db *sqlx.DB, accountID, userID, entityID, itemID string, newFields, oldFields map[string]interface{}) *Message {
	m := &Message{
		ID:        fmt.Sprintf("%s#%s", "update", uuid.New().String()),
		Type:      TypeItemUpdate,
		AccountID: accountID,
		UserID:    userID,
		EntityID:  entityID,
		ItemID:    itemID,
		NewFields: newFields,
		OldFields: oldFields,
		State:     StateQueued,
	}

	add(ctx, db, m, "Queued", StateQueued)
	return m
}

func NewDeleteItemMessage(ctx context.Context, db *sqlx.DB, accountID, userID, entityID, itemID string) *Message {
	m := &Message{
		ID:        fmt.Sprintf("%s#%s", "delete", uuid.New().String()),
		Type:      TypeItemDelete,
		AccountID: accountID,
		UserID:    userID,
		EntityID:  entityID,
		ItemID:    itemID,
		State:     StateQueued,
	}

	add(ctx, db, m, "Queued", StateQueued)
	return m
}

func NewReminderMessage(ctx context.Context, db *sqlx.DB, accountID, userID, entityID, itemID string) *Message {
	m := &Message{
		ID:        fmt.Sprintf("%s#%s", "reminder", uuid.New().String()),
		Type:      TypeItemRemind,
		AccountID: accountID,
		UserID:    userID,
		EntityID:  entityID,
		ItemID:    itemID,
		State:     StateQueued,
	}

	add(ctx, db, m, "Queued", StateQueued)
	return m
}

func NewDelayMessage(ctx context.Context, db *sqlx.DB, accountID, userID, entityID, itemID string, meta map[string]interface{}) *Message {
	m := &Message{
		ID:        fmt.Sprintf("%s#%s", "delay", uuid.New().String()),
		Type:      TypeItemDelayed,
		AccountID: accountID,
		UserID:    userID,
		EntityID:  entityID,
		ItemID:    itemID,
		Meta:      meta,
		State:     StateQueued,
	}
	add(ctx, db, m, "Queued", StateQueued)
	return m
}

func NewEmailConversationMessage(ctx context.Context, db *sqlx.DB, accountID, userID, entityID, itemID, conversationID string, source map[string][]string) *Message {
	m := &Message{
		ID:             fmt.Sprintf("%s#%s", "emailconversation", uuid.New().String()),
		Type:           TypeEmailConversationAdded,
		AccountID:      accountID,
		UserID:         userID,
		EntityID:       entityID,
		ItemID:         itemID,
		ConversationID: conversationID,
		Source:         source,
		State:          StateQueued,
	}
	add(ctx, db, m, "Queued", StateQueued)
	return m
}

func NewChatConversationMessage(ctx context.Context, db *sqlx.DB, accountID, userID, entityID, itemID, conversationID string, source map[string][]string) *Message {
	m := &Message{
		ID:             fmt.Sprintf("%s#%s", "chatconversation", uuid.New().String()),
		Type:           TypeChatConversationAdded,
		AccountID:      accountID,
		UserID:         userID,
		EntityID:       entityID,
		ItemID:         itemID,
		ConversationID: conversationID,
		Source:         source,
		State:          StateQueued,
	}
	add(ctx, db, m, "Queued", StateQueued)
	return m
}

func (m Message) TypeStr() string {
	switch m.Type {
	case TypeDefault:
		return "Type Default"
	case TypeItemCreate:
		return "Type Create"
	case TypeItemUpdate:
		return "Type Update"
	case TypeItemDelete:
		return "Type Delete"
	case TypeItemRemind:
		return "Type Remind"
	case TypeItemDelayed:
		return "Type Delay"
	case TypeEmailConversationAdded:
		return "Type Email Conversation"
	case TypeChatConversationAdded:
		return "Type Chat Conversation"
	default:
		return "Type Not Implemented"
	}
}
