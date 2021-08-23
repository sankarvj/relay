package conversation

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// Create add new conversation with respective types.
func Create(ctx context.Context, db *sqlx.DB, nc NewConversation, now time.Time) (Conversation, error) {
	ctx, span := trace.StartSpan(ctx, "internal.conversation.Create")
	defer span.End()

	payloadBytes, err := json.Marshal(nc.Payload)
	if err != nil {
		return Conversation{}, errors.Wrap(err, "encode payload to bytes")
	}

	c := Conversation{
		ID:        nc.ID,
		AccountID: nc.AccountID,
		EntityID:  nc.EntityID,
		ItemID:    nc.ItemID,
		UserID:    nc.UserID,
		Type:      nc.Type,
		Message:   nc.Message,
		Payload:   string(payloadBytes),
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC().Unix(),
	}

	const q = `INSERT INTO conversations
		(conversation_id, account_id, entity_id, item_id, user_id, type, message, payload, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err = db.ExecContext(
		ctx, q,
		c.ID, c.AccountID, c.EntityID, c.ItemID, c.UserID, c.Type, c.Message, c.Payload,
		c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return Conversation{}, errors.Wrap(err, "inserting conversation")
	}

	return c, nil
}

func List(ctx context.Context, accountID, entityID, itemID string, ctype int, db *sqlx.DB) ([]ViewModelConversation, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.List")
	defer span.End()

	conversations := []ViewModelConversation{}
	const q = `SELECT c.conversation_id as conversation_id, c.user_id as user_id, u.name as user_name, u.avatar as user_avatar, c.type as type, c.message as message FROM conversations as c join users as u on u.user_id = c.user_id where c.account_id = $1 AND c.entity_id = $2 AND c.item_id = $3`

	if err := db.SelectContext(ctx, &conversations, q, accountID, entityID, itemID); err != nil {
		return nil, errors.Wrap(err, "selecting conversations")
	}

	return conversations, nil
}
