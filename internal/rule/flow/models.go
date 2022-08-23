package flow

import (
	"time"

	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

//FlowType is the type of flow
const (
	FlowTypeAll           = -1
	FlowTypeUnknown       = 0
	FlowTypeEntersSegment = 1
	FlowTypeLeavesSegment = 2
	FlowTypeEventCreate   = 3
	FlowTypeEventUpdate   = 4
)

const (
	FlowModeAll      = -1
	FlowModeWorkFlow = 0
	FlowModePipeLine = 1
	FlowModeSegment  = 2
)

const (
	FlowStatusActive   = 0
	FlowStatusInActive = 1
)

// FlowCondition defines exists/entry conditions
//it will be used to identify whether to hold or continue the execution
const (
	FlowConditionNil   = -1
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
	Tokenb      *string   `db:"tokenb" json:"tokenb"` //useful for displaying the values in the filterconditions.ts
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Mode        int       `db:"mode" json:"mode"`
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
	ID          string                 `json:"id"`
	EntityID    string                 `json:"entity_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Expression  string                 `json:"expression"`
	Mode        int                    `json:"mode"`
	Type        int                    `json:"type"`
	Status      int                    `json:"status"`
	Nodes       []node.ViewModelNode   `json:"nodes"`
	Tokens      map[string]interface{} `json:"tokens"`
}

// NewFlow has information needed to creat new flow
type NewFlow struct {
	ID          string                 `json:"id"`
	AccountID   string                 `json:"account_id"`
	EntityID    string                 `json:"entity_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Expression  string                 `json:"expression"`
	Tokens      map[string]interface{} `json:"tokens"`
	Mode        int                    `json:"mode"`
	Type        int                    `json:"type"`
	Status      int                    `json:"status"`
	Condition   int                    `json:"condition"`
	Nodes       []node.NewNode         `json:"nodes"`
	Queries     []node.Query           `json:"queries"`
}

// ActiveNode represents the node which are currently active
type ActiveNode struct {
	AccountID string `db:"account_id" json:"account_id"`
	FlowID    string `db:"flow_id" json:"flow_id"`
	EntityID  string `db:"entity_id" json:"entity_id"`
	ItemID    string `db:"item_id" json:"item_id"`
	NodeID    string `db:"node_id" json:"node_id"`
	Life      int    `db:"life" json:"life"`
	IsActive  bool   `db:"is_active" json:"is_active"`
}
