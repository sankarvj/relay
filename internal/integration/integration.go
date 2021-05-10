package integration

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Actions interface {
	Act(ctx context.Context, accountID, actionID string, actionPayload ActionPayload, db *sqlx.DB) error
}

type ActionPayload struct {
	ID      string            `json:"id"`
	Payload map[string]string `json:"payload"`
}
