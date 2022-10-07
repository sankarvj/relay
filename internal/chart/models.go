package chart

import (
	"time"
)

const (
	MetaSourceKey    = "source"
	MetaFieldKey     = "field"
	MetaDataType     = "data_type"
	MetaCalcKey      = "calc"
	MetaGroupByLogic = "group_by_logic"
	MetaExp          = "exp"
	MetaDateField    = "date"
)

type Calc string

const (
	CalcRate  Calc = "rate"
	CalcCount Calc = "count"
	CalcSum   Calc = "sum"
)

type GroupLogic string

const (
	GroupLogicNone   GroupLogic = "none"
	GroupLogicID     GroupLogic = "g_b_id"
	GroupLogicField  GroupLogic = "g_b_f"
	GroupLogicParent GroupLogic = "g_b_p"
)

type Type string

const (
	TypePie  Type = "pie"
	TypeLine Type = "line"
	TypeBar  Type = "bar"
	TypeGrid Type = "grid"
	TypeRod  Type = "rod"
)

type DType string

const (
	DTypeTimeseries DType = "timeseries"
	DTypeDefault    DType = "default"
)

type Duration string

const (
	AllTime   Duration = "all_time"
	Last24Hrs Duration = "last_24hrs"
	LastWeek  Duration = "last_week"
)

type Chart struct {
	ID        string    `db:"chart_id" json:"id"`
	AccountID string    `db:"account_id" json:"account_id"`
	EntityID  string    `db:"entity_id" json:"entity_id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Name      string    `db:"name" json:"name"`
	Type      string    `db:"type" json:"type"`
	Duration  string    `db:"duration" json:"duration"`
	State     int       `db:"state" json:"state"`
	Position  int       `db:"position" json:"position"`
	Metab     string    `db:"metab" json:"metab"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type NewChart struct {
	AccountID string            `json:"account_id"`
	EntityID  string            `json:"entity_id"`
	UserID    string            `json:"user_id"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Duration  string            `json:"duration"`
	State     int               `json:"state"`
	Position  int               `json:"position"`
	Meta      map[string]string `json:"meta"`
}
