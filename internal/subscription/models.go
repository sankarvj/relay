package subscription

import "time"

type Subscription struct {
	ID        string    `db:"subscription_id" json:"id"`
	AccountID string    `db:"account_id" json:"account_id"`
	EntityID  string    `db:"entity_id" json:"entity_id"`
	ItemID    string    `db:"item_id" json:"item_id"`
	UserID    string    `db:"user_id" json:"user_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt int64     `db:"updated_at" json:"updated_at"`
}

type NewSubscription struct {
	ID        string `json:"id"`
	AccountID string `json:"account_id"`
	EntityID  string `json:"entity_id"`
	ItemID    string `json:"item_id"`
	UserID    string `json:"user_id"`
}
