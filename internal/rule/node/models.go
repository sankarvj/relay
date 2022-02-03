package node

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"
)

//Type specifies the enum for type of nodes

//consts for different node types
const (
	Unknown  int = 0
	Decision     = 1
	Push         = 2
	Modify       = 3
	Hook         = 5
	Schedule     = 6
	Delay        = 7
	Stage        = 8
	//specific types equivalent to push
	Task    = 101
	Meeting = 102
	Email   = 103
)

//Node struct defines the structure of each node in the workflow
type Node struct {
	ID           string    `db:"node_id" json:"id"`
	ParentNodeID string    `db:"parent_node_id" json:"parent_node_id"` //put 000000 for default
	AccountID    string    `db:"account_id" json:"account_id"`
	FlowID       string    `db:"flow_id" json:"flow_id"`
	ActorID      string    `db:"actor_id" json:"actor_id"`
	StageID      string    `db:"stage_id" json:"stage_id"`
	Name         string    `db:"name" json:"name"`
	Description  string    `db:"description" json:"description"`
	Weight       int       `db:"weight" json:"weight"`
	Type         int       `db:"type" json:"type"`
	Expression   string    `db:"expression" json:"expression"`
	Tokenb       string    `db:"tokenb" json:"tokenb"`   //useful for displaying the values in the filterconditions.ts
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
	FlowType int
}

//Node struct defines the structure of each node in the workflow
type NodeActor struct {
	ID             string         `db:"node_id" json:"id"`
	FlowID         string         `db:"flow_id" json:"flow_id"`
	ParentNodeID   string         `db:"parent_node_id" json:"parent_node_id"`
	StageID        string         `db:"stage_id" json:"stage_id"`
	Name           string         `db:"name" json:"name"`
	Description    string         `db:"description" json:"description"`
	Weight         int            `db:"weight" json:"weight"`
	ActorID        string         `db:"actor_id" json:"actor_id"`
	EntityName     sql.NullString `db:"entity_name" json:"entity_name"`
	EntityCategory sql.NullInt32  `db:"category" json:"category"`
	Type           int            `db:"type" json:"type"`
	Expression     string         `db:"expression" json:"expression"`
	Tokenb         string         `db:"tokenb" json:"tokenb"` //useful for displaying the values in the filterconditions.ts
	Actuals        string         `db:"actuals" json:"actuals"`
}

// ViewModelNode represents the view model of node
type ViewModelNode struct {
	ID             string                 `json:"id"`
	FlowID         string                 `json:"flow_id"`
	StageID        string                 `json:"stage_id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Weight         int                    `json:"weight"`
	Expression     string                 `json:"expression"`
	ParentNodeID   string                 `json:"parent_node_id"`
	EntityName     string                 `json:"entity_name"`
	EntityCategory int                    `json:"entity_category"`
	ActorID        string                 `json:"actor_id"`
	Type           int                    `json:"type"`
	Actuals        map[string]string      `json:"actuals"`
	Exp            string                 `json:"exp"`
	Tokens         map[string]interface{} `json:"tokens"`
}

type ViewModelActiveNode struct {
	ID       string `json:"id"`
	IsActive bool   `json:"is_active"`
	Life     int    `json:"life"`
}

// NewNode has information needed to creat new node
type NewNode struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Weight       int                    `json:"weight"`
	ParentNodeID string                 `json:"parent_node_id"`
	AccountID    string                 `json:"account_id"`
	FlowID       string                 `json:"flow_id"`
	ActorID      string                 `json:"actor_id"`
	StageID      string                 `json:"stage_id"`
	Type         int                    `json:"type" validate:"required"`
	Expression   string                 `json:"expression"`
	Tokens       map[string]interface{} `json:"tokens"`
	Actuals      map[string]string      `json:"actuals"`
	Queries      []Query                `json:"queries"`
}

type Query struct {
	EntityID string `json:"entity_id"`
	Key      string `json:"key"`
	Value    string `json:"value"`
	Operator string `json:"operator"`
	Display  string `json:"display"`
}

type NodeMapWrapper struct {
	Mapper map[string]string `json:"mapper"`
}

// VariablesMap parses variables jsonb to map
func (n Node) VariablesMap() map[string]interface{} {
	var variables map[string]interface{}
	if err := json.Unmarshal([]byte(n.Variables), &variables); err != nil && n.Variables != "" {
		log.Printf("critical error occurred when unmarshalling node variables %v %v\n", n.ID, err)
		panic(err)
	}
	return variables
}

func (n Node) VarStrMap() map[string]string {
	varStrMap := make(map[string]string, 0)
	vars := n.VariablesMap()
	for k, v := range vars {
		if k == GlobalEntity {
			log.Println("rule.node.models: TODO What can be done here? shall we convert map to string?")
		} else {
			varStrMap[k] = v.(string)
		}

	}
	return varStrMap
}

// ActualsMap parses actuals jsonb to map
func (n Node) ActualsMap() map[string]string {
	var actuals map[string]string
	if err := json.Unmarshal([]byte(n.Actuals), &actuals); err != nil {
		log.Printf("critical error occurred while unmarshalling node actuals %v %v\n", n.ID, err)
		panic(err)
	}
	return actuals
}

// ActualsMap parses actuals jsonb to map
func (n Node) ActualsItemID() string {
	return n.ActualsMap()[n.ActorID]
}
