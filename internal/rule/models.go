package rule

import "time"

// Rule represents an rule item for the entity
type Rule struct {
	ID         string    `db:"rule_id" json:"id"`
	EntityID   string    `db:"entity_id" json:"entity_id"`
	Expression string    `db:"expression" json:"expression"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  int64     `db:"updated_at" json:"updated_at"`
}

// ViewModelRule represents the view model of rule
// (i.e) it has fields instead of attributes
type ViewModelRule struct {
	ID         string `json:"id"`
	EntityID   string `json:"entity_id"`
	Expression string `json:"expression"`
}

// NewRule has information needed to creat new rule
type NewRule struct {
	EntityID   string    `json:"entity_id"`
	Expression string    `json:"expression"`
	Action     NewAction `json:"action"`
}

// NewAction has information needed to creat new action
type NewAction struct {
	EntityID    string            `json:"entity_id"`
	Set         map[string]string `json:"set"`
	Condition   map[string]string `json:"condition"`
	Uncondition map[string]string `json:"uncondition"`
}
