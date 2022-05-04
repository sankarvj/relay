package stream

//Type for the item is still open we can use it for anything
const (
	TypeDefault           = 0
	TypeItemCreate        = 1
	TypeItemUpdate        = 2
	TypeItemDelete        = 3
	TypeItemRemind        = 4
	TypeItemDelayed       = 5
	TypeConversationAdded = 6
)

type Message struct {
	Type           int                    `json:"type"`
	AccountID      string                 `json:"account_id"`
	UserID         string                 `json:"user_id"`
	EntityID       string                 `json:"entity_id"`
	ItemID         string                 `json:"item_id"`
	ConversationID string                 `json:"conversation_id"`
	NewFields      map[string]interface{} `json:"new_fields"`
	OldFields      map[string]interface{} `json:"old_fields"`
	Source         map[string]string      `json:"source"`
	Meta           map[string]interface{} `json:"meta"`
}

func NewCreteItemMessage(accountID, userID, entityID, itemID string, source map[string]string) *Message {
	return &Message{
		Type:      TypeItemCreate,
		AccountID: accountID,
		UserID:    userID,
		EntityID:  entityID,
		ItemID:    itemID,
		Source:    source,
	}
}

func NewUpdateItemMessage(accountID, userID, entityID, itemID string, newFields, oldFields map[string]interface{}) *Message {
	return &Message{
		Type:      TypeItemUpdate,
		AccountID: accountID,
		UserID:    userID,
		EntityID:  entityID,
		ItemID:    itemID,
		NewFields: newFields,
		OldFields: oldFields,
	}
}

func NewDeleteItemMessage(accountID, userID, entityID, itemID string) *Message {
	return &Message{
		Type:      TypeItemDelete,
		AccountID: accountID,
		UserID:    userID,
		EntityID:  entityID,
		ItemID:    itemID,
	}
}

func NewReminderMessage(accountID, userID, entityID, itemID string) *Message {
	return &Message{
		Type:      TypeItemRemind,
		AccountID: accountID,
		UserID:    userID,
		EntityID:  entityID,
		ItemID:    itemID,
	}
}

func NewDelayMessage(accountID, userID, entityID, itemID string, meta map[string]interface{}) *Message {
	return &Message{
		Type:      TypeItemDelayed,
		AccountID: accountID,
		UserID:    userID,
		EntityID:  entityID,
		ItemID:    itemID,
		Meta:      meta,
	}
}

func NewConversationMessage(accountID, userID, entityID, itemID, conversationID string) *Message {
	return &Message{
		Type:           TypeConversationAdded,
		AccountID:      accountID,
		UserID:         userID,
		EntityID:       entityID,
		ItemID:         itemID,
		ConversationID: conversationID,
	}
}
