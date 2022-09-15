package connection

import "time"

type Connection struct {
	ConnectionID   string    `db:"connection_id" json:"connection_id"`
	RelationshipID string    `db:"relationship_id" json:"relationship_id"`
	AccountID      string    `db:"account_id" json:"account_id"`
	UserID         string    `db:"user_id" json:"user_id"`
	EntityName     string    `db:"entity_name" json:"entity_name"`
	SrcEntityID    string    `db:"src_entity_id" json:"src_entity_id"`
	DstEntityID    string    `db:"dst_entity_id" json:"dst_entity_id"`
	SrcItemID      string    `db:"src_item_id" json:"src_item_id"`
	DstItemID      string    `db:"dst_item_id" json:"dst_item_id"`
	Title          string    `db:"title" json:"title"`
	SubTitle       string    `db:"sub_title" json:"sub_title"`
	Action         string    `db:"action" json:"action"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      int64     `db:"updated_at" json:"updated_at"`
}

type ConnectedItem struct {
	DstItemID    string    `db:"dst_item_id"  json:"dst_item_id"`
	ConnectionID string    `db:"connection_id"  json:"connection_id"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
