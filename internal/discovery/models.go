package discovery

import "time"

type Discover struct {
	ID        string    `db:"discovery_id" json:"id"`
	Type      string    `db:"discovery_type" json:"type"`
	AccountID string    `db:"account_id" json:"account_id"`
	EntityID  string    `db:"entity_id" json:"entity_id"`
	ItemID    string    `db:"item_id" json:"item_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt int64     `db:"updated_at" json:"updated_at"`
}

type NewDiscovery struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	AccountID string `json:"account_id"`
	EntityID  string `json:"entity_id"`
	ItemID    string `json:"item_id"`
}
