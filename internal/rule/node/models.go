package node

import (
	"encoding/json"
	"log"
	"time"
)

//Type specifies the enum for type of nodes

//consts for different node types
const (
	Decision int = 0
	Push         = 1
	Modify       = 2
	Email        = 3
	Hook         = 4
	Schedule     = 5
	Delay        = 6
)

//Node struct defines the structure of each node in the workflow
type Node struct {
	ID           string    `db:"node_id" json:"id"`
	ParentNodeID *string   `db:"parent_node_id" json:"parent_node_id"`
	AccountID    string    `db:"account_id" json:"account_id"`
	FlowID       string    `db:"flow_id" json:"flow_id"`
	ActorID      string    `db:"actor_id" json:"actor_id"`
	Type         int       `db:"type" json:"type"`
	Expression   string    `db:"expression" json:"expression"`
	Variables    string    `db:"-" json:"variables"`     //Variables is to evaluate the expression during the runtime
	Meta         Meta      `db:"-" json:"meta"`          //Meta used to pass meta data to queues during specific node run time
	Actuals      string    `db:"actuals" json:"actuals"` //Actuals used to get the actual item of the actionable entity - ActorID's item
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    int64     `db:"updated_at" json:"updated_at"`
}

//Meta is the node meta data which is passed to the queues
type Meta struct {
	EntityID string
	ItemID   string
}

// ViewModelNode represents the view model of node
type ViewModelNode struct {
	ID           string `json:"id"`
	ParentNodeID string `json:"parent_node_id"`
	Type         int    `json:"type"`
}

// NewNode has information needed to creat new node
type NewNode struct {
	AccountID    string            `json:"account_id" validate:"required"`
	ParentNodeID string            `json:"parent_node_id"`
	FlowID       string            `json:"flow_id" validate:"required"`
	ActorID      string            `json:"actor_id"`
	Type         int               `json:"type" validate:"required"`
	Expression   string            `json:"expression"`
	Actuals      map[string]string `json:"actuals"`
}

// VariablesMap parses variables jsonb to map
func (n Node) VariablesMap() map[string]interface{} {
	var variables map[string]interface{}
	if err := json.Unmarshal([]byte(n.Variables), &variables); err != nil {
		log.Printf("error while unmarshalling node variables %v %v", n.ID, err)
		panic(err)
	}
	return variables
}

// ActualsMap parses actuals jsonb to map
func (n Node) ActualsMap() map[string]string {
	var actuals map[string]string
	if err := json.Unmarshal([]byte(n.Actuals), &actuals); err != nil {
		log.Printf("error while unmarshalling node actuals %v %v", n.ID, err)
		panic(err)
	}
	return actuals
}
