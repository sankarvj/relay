package stream

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific entity is requested but does not exist.
	ErrNotFound = errors.New("Log Stream not found")
)

type LogStream struct {
	AccountID string    `db:"account_id" json:"account_id"`
	LogID     string    `db:"log_id" json:"log_id"`
	Comment   string    `db:"comment" json:"comment"`
	State     int       `db:"state" json:"state"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

func Retrieve(ctx context.Context, accountID, logID string, db *sqlx.DB) (LogStream, error) {
	ctx, span := trace.StartSpan(ctx, "internal.log.Retrieve")
	defer span.End()

	var ls LogStream
	const q = `SELECT * FROM log_streams WHERE account_id = $1 AND log_id = $2`
	if err := db.GetContext(ctx, &ls, q, accountID, logID); err != nil {
		if err == sql.ErrNoRows {
			return LogStream{}, ErrNotFound
		}

		return LogStream{}, errors.Wrapf(err, "selecting log stream %q", logID)
	}

	return ls, nil
}

func Add(ctx context.Context, db *sqlx.DB, m *Message, comment string, state int) {
	ctx, span := trace.StartSpan(ctx, "internal.log.Add")
	defer span.End()

	m.Comment = comment
	m.State = state

	const q = `INSERT INTO log_streams
		(account_id, log_id, comment, state, created_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := db.ExecContext(
		ctx, q,
		m.AccountID, m.ID, comment, state,
		time.Now().UTC(),
	)
	if err != nil {
		log.Println("unexpected error when adding log", err)
	}
}

func Update(ctx context.Context, db *sqlx.DB, m *Message, comment string, state int) {
	ctx, span := trace.StartSpan(ctx, "internal.log.Update")
	defer span.End()

	log.Printf("log message updated for :: %s  -- comment :: %s  --  state :: %d", m.ID, m.Comment, m.State)

	m.Comment = comment
	m.State = state

	const q = `UPDATE log_streams SET
				"comment" = $3,
				"state" = $4
				WHERE account_id = $1 AND log_id = $2`
	_, err := db.ExecContext(ctx, q, m.AccountID, m.ID,
		comment, state)
	if err != nil {
		log.Println("unexpected error when updating log state", err)
	}
}

func Delete(ctx context.Context, db *sqlx.DB, accountID, logID string) {
	ctx, span := trace.StartSpan(ctx, "internal.log.Delete")
	defer span.End()

	const q = `DELETE FROM log_streams WHERE account_id = $1 and log_id = $2`

	if _, err := db.ExecContext(ctx, q, accountID, logID); err != nil {
		log.Println("unexpected error when deleting log", err)
	}
}
