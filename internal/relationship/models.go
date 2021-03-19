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
	DisplayName    string `db:"display_name" json:"display_name"`
	Category       int    `db:"category" json:"category"`
	EntityID       string `db:"entity_id" json:"entity_id"`
	Type           RType  `db:"type" json:"type"`
}

type Relatable struct {
	RefID string
	RType RType
}

//RType defines the type of relationships
type RType int

//Relationships could be defind in the fields itself (ex: deals has associated contacts field) or
//Relationships could be defind explicitly (ex: deals has tickets and tickets has deals but not exposed explicitly)
const (
	RTypeBothSide RType = 0
	RTypeSrcSide        = 1
	RTypeDstSide        = 2
)

const (
	FieldAssociationKey string = "00000000-0000-0000-0000-000000000000"
)

func MakeRelatable(refID string, rtype RType) Relatable {
	return Relatable{
		RefID: refID,
		RType: rtype,
	}
}
