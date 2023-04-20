package timeseries

import (
	"time"

	"github.com/lib/pq"
)

const (
	TypeUnknown  = 0
	TypeIncident = 1
)

type Timeseries struct {
	ID          string         `db:"timeseries_id" json:"timeseries_id"`
	AccountID   string         `db:"account_id" json:"account_id"`
	EntityID    string         `db:"entity_id" json:"entity_id"`
	Identifier  *string        `db:"identifier" json:"identifier"` // * because it could be null
	Type        int            `db:"type" json:"type"`
	Event       string         `db:"event" json:"event"`
	Description string         `db:"description" json:"description"`
	Count       int            `db:"count" json:"count"`
	Tags        pq.StringArray `db:"tags" json:"tags"`
	Fieldsb     string         `db:"fieldsb" json:"fieldsb"`
	StartTime   time.Time      `db:"start_time" json:"start_time"`
	EndTime     time.Time      `db:"end_time" json:"end_time"`
}

type NewTimeseries struct {
	ID          string                 `json:"timeseries_id"`
	AccountID   string                 `json:"account_id"`
	EntityID    string                 `json:"entity_id"`
	Identifier  *string                `json:"identifier"`
	Type        int                    `json:"type"`
	Event       string                 `json:"event"`
	Description string                 `json:"description"`
	Count       int                    `json:"count"`
	Tags        []string               `json:"tags"`
	Fields      map[string]interface{} `json:"fields"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
}
