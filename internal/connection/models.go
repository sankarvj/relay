package connection

type Connection struct {
	RelationshipID string `db:"relationship_id" json:"relationship_id"`
	AccountID      string `db:"account_id" json:"account_id"`
	SrcItemID      string `db:"src_item_id" json:"src_item_id"`
	DstItemID      string `db:"dst_item_id" json:"dst_item_id"`
}

type ConnectionID struct {
	SrcItemID string `db:"src_item_id" json:"src_item_id"`
	DstItemID string `db:"dst_item_id" json:"dst_item_id"`
}
