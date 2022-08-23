package conversation

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/schema"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific entity is requested but does not exist.
	ErrNotFound = errors.New("Conversation not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("Conversation ID is not in its proper form")
)

func SaveConversation(ctx context.Context, accountID, entityID, parentItemId string, emailEntityItem entity.EmailEntity, message string, convType, convState int, db *sqlx.DB) error {
	newConversation := NewConversation{
		ID:        emailEntityItem.MessageID,
		AccountID: accountID,
		EntityID:  entityID,
		ItemID:    &parentItemId,
		UserID:    schema.SeedSystemUserID,
		Message:   message,
		Type:      convType,
		State:     convState,
		Payload:   util.ConvertInterfaceToMap(emailEntityItem),
	}
	_, err := Create(ctx, db, newConversation, time.Now())
	return err
}

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
		State:     nc.State,
		Message:   nc.Message,
		Payload:   string(payloadBytes),
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC().Unix(),
	}

	const q = `INSERT INTO conversations
		(conversation_id, account_id, entity_id, item_id, user_id, type, state, message, payload, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = db.ExecContext(
		ctx, q,
		c.ID, c.AccountID, c.EntityID, c.ItemID, c.UserID, c.Type, c.State, c.Message, c.Payload,
		c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return Conversation{}, errors.Wrap(err, "inserting conversation")
	}

	return c, nil
}

func List(ctx context.Context, accountID, entityID, itemID string, ctype int, db *sqlx.DB) ([]ConversationUsr, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.List")
	defer span.End()

	conversations := []ConversationUsr{}
	const q = `SELECT c.conversation_id as conversation_id, c.user_id as user_id, u.name as user_name, u.avatar as user_avatar, c.type as type, c.state as state, c.message as message, c.created_at as created_at FROM conversations as c left join users as u on u.user_id = c.user_id where c.account_id = $1 AND c.entity_id = $2 AND c.item_id = $3 ORDER BY c.created_at ASC LIMIT 500`

	if err := db.SelectContext(ctx, &conversations, q, accountID, entityID, itemID); err != nil {
		return nil, errors.Wrap(err, "selecting conversations")
	}

	return conversations, nil
}

func Retrieve(ctx context.Context, accountID, conversationID string, db *sqlx.DB) (Conversation, error) {
	ctx, span := trace.StartSpan(ctx, "internal.conversation.Retrieve")
	defer span.End()

	var cv Conversation
	const q = `SELECT * FROM conversations WHERE account_id = $1 AND conversation_id = $2`
	if err := db.GetContext(ctx, &cv, q, accountID, conversationID); err != nil {
		if err == sql.ErrNoRows {
			return Conversation{}, ErrNotFound
		}

		return Conversation{}, errors.Wrapf(err, "selecting conversation %q", conversationID)
	}

	return cv, nil
}

func UpdateID(ctx context.Context, db *sqlx.DB, convID, newConvID string, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.conversation.UpdateID")
	defer span.End()

	const q = `UPDATE conversations SET
		"conversation_id" = $2,
		"updated_at" = $3
		WHERE conversation_id = $1`
	_, err := db.ExecContext(ctx, q, convID, newConvID,
		now.Unix(),
	)
	if err != nil {
		return errors.Wrap(err, "updating conversation")
	}

	return nil
}

func (c Conversation) PayloadMap() map[string]interface{} {
	var payload map[string]interface{}
	if c.Payload == "" {
		return payload
	}
	if err := json.Unmarshal([]byte(c.Payload), &payload); err != nil {
		log.Printf("***> unexpected error occurred when unmarshalling payload for conversation: %v error: %v\n", c.ID, err)
	}
	return payload
}
