package flow

import (
	"time"

	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

//FlowType is the type of flow
const (
	FlowTypeFieldUpdate = 0
	FlowTypeSegment     = 1
	FlowTypeEvent       = 2
	FlowTypePipeline    = 3
)

//FlowCondition defines exists/entry conditions
const (
	FlowConditionBoth  = 0
	FlowConditionEntry = 1
	FlowConditionExit  = 2
)

// Flow represents a single workflow
type Flow struct {
	ID          string    `db:"flow_id" json:"id"`
	AccountID   string    `db:"account_id" json:"account_id"`
	EntityID    string    `db:"entity_id" json:"entity_id"`
	Expression  string    `db:"expression" json:"expression"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Type        int       `db:"type" json:"type"`
	Condition   int       `db:"condition" json:"condition"`
	Status      int       `db:"status" json:"status"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   int64     `db:"updated_at" json:"updated_at"`
}

// ActiveFlow represents the flow which are currently active
type ActiveFlow struct {
	AccountID string `db:"account_id" json:"account_id"`
	FlowID    string `db:"flow_id" json:"flow_id"`
	ItemID    string `db:"item_id" json:"item_id"`
	NodeID    string `db:"node_id" json:"node_id"`
	Life      int    `db:"life" json:"life"`
	IsActive  bool   `db:"is_active" json:"is_active"`
}

// ViewModelFlow represents the view model of flow
type ViewModelFlow struct {
	ID          string `json:"id"`
	EntityID    string `json:"entity_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Expression  string `json:"condition"`
}

// NewFlow has information needed to creat new flow
type NewFlow struct {
	ID          string         `json:"id"`
	AccountID   string         `json:"account_id"`
	EntityID    string         `json:"entity_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Expression  string         `json:"expression"`
	Type        int            `json:"type"`
	Condition   int            `json:"condition"`
	Nodes       []node.NewNode `json:"nodes"`
}

// ActiveNode represents the node which are currently active
type ActiveNode struct {
	AccountID string `db:"account_id" json:"account_id"`
	FlowID    string `db:"flow_id" json:"flow_id"`
	ItemID    string `db:"item_id" json:"item_id"`
	NodeID    string `db:"node_id" json:"node_id"`
	Life      int    `db:"life" json:"life"`
	IsActive  bool   `db:"is_active" json:"is_active"`
}
