package token

import (
	"time"

	"github.com/lib/pq"
)

type Token struct {
	Token     string         `db:"token" json:"token"`
	AccountID string         `db:"account_id" json:"account_id"`
	Type      int            `db:"type" json:"type"`
	State     int            `db:"state" json:"state"`
	Scope     pq.StringArray `db:"scope" json:"scope"`
	IssuedAt  time.Time      `db:"issued_at" json:"issued_at"`
	Expiry    time.Time      `db:"expiry" json:"expiry"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
}
