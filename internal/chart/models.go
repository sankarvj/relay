package chart

import (
	"time"
)

const (
	CalcCount = 0
	CalcSum   = 1
)

type Chart struct {
	ID             string    `db:"chart_id" json:"id"`
	AccountID      string    `db:"account_id" json:"account_id"`
	EntityID       string    `db:"entity_id" json:"entity_id"`
	ParentEntityID string    `db:"parent_entity_id" json:"parent_entity_id"`
	UserID         string    `db:"user_id" json:"user_id"`
	Name           string    `db:"name" json:"name"`
	Field          string    `db:"field" json:"field"`
	Type           string    `db:"type" json:"type"`
	Group          string    `db:"group" json:"group"`
	Duration       string    `db:"duration" json:"duration"`
	State          int       `db:"state" json:"state"`
	Calc           int       `db:"calc" json:"calc"`
	Position       int       `db:"position" json:"position"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

type NewChart struct {
	AccountID      string `json:"account_id"`
	EntityID       string `json:"entity_id"`
	ParentEntityID string `json:"parent_entity_id"`
	UserID         string `json:"user_id"`
	Name           string `json:"name"`
	Field          string `json:"field"`
	Type           string `json:"type"`
	Group          string `json:"group"`
	Duration       string `json:"duration"`
	State          int    `json:"state"`
	Calc           int    `json:"calc"`
	Position       int    `json:"position"`
}
