package relationship

// Relationship represents relationship of the reference fields
type Relationship struct {
	RelationshipID string `db:"relationship_id" json:"relationship_id"`
	AccountID      string `db:"account_id" json:"account_id"`
	SrcEntityID    string `db:"src_entity_id" json:"src_entity_id"`
	DstEntityID    string `db:"dst_entity_id" json:"dst_entity_id"`
	FieldID        string `db:"field_id" json:"field_id"`
	Type           RType  `db:"type" json:"type"`
}

type Bond struct {
	RelationshipID string `db:"relationship_id" json:"relationship_id"`
	Name           string `db:"name" json:"name"`
	Category       int    `db:"category" json:"category"`
	EntityID       string `db:"entity_id" json:"entity_id"`
	Type           RType  `db:"type" json:"type"`
}

//RType defines the type of relationships
type RType int

//Mode for the entity spcifies certain entity specific characteristics
//Keep this as minimal and add a sub-type for data types such as decimal,boolean,time & date
const (
	TypeBond         RType = 0
	TypeAssociation        = 1
	TypeImplicitBond       = 2 // useful for bond like one-to-many associations but not as the field property. contact-activities
)

const (
	FieldAssociationKey string = "00000000-0000-0000-0000-000000000000"
)
