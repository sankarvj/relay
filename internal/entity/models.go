package entity

import (
	"time"

	"github.com/lib/pq"
)

// Entity represents the building block of all the tasks
type Entity struct {
	ID            string         `db:"entity_id" json:"id"`
	AccountID     string         `db:"account_id" json:"account_id"`
	TeamID        string         `db:"team_id" json:"team_id"`
	Name          string         `db:"name" json:"name"`
	DisplayName   string         `db:"display_name" json:"display_name"`
	Category      int            `db:"category" json:"category"`
	State         int            `db:"state" json:"state"`
	Status        int            `db:"status" json:"status"`
	Retry         int            `db:"retry" json:"retry"`
	Fieldsb       string         `db:"fieldsb" json:"fieldsb"`
	Metab         *string        `db:"metab" json:"metab"`
	Tags          pq.StringArray `db:"tags" json:"tags"`
	IsPublic      bool           `db:"is_public" json:"is_public"`
	IsCore        bool           `db:"is_core" json:"is_core"`
	IsShared      bool           `db:"is_shared" json:"is_shared"`
	SharedTeamIds pq.StringArray `db:"shared_team_ids" json:"shared_team_ids"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt     int64          `db:"updated_at" json:"updated_at"`
}

// ViewModelEntity represents the view model of entity
// (i.e) it has fields instead of attributes
type ViewModelEntity struct {
	ID            string    `json:"id"`
	TeamID        string    `json:"team_id"`
	Name          string    `json:"name"`
	DisplayName   string    `json:"display_name"`
	Category      int       `json:"category"`
	State         int       `json:"state"`
	Status        int       `json:"status"`
	Fields        []Field   `json:"fields"`
	Tags          []string  `json:"tags"`
	IsPublic      bool      `json:"is_public"`
	IsCore        bool      `json:"is_core"`
	IsShared      bool      `json:"is_shared"`
	SharedTeamIds []string  `json:"shared_team_ids"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     int64     `json:"updated_at"`
}

// NewEntity has information needed to creat new entity
type NewEntity struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name" validate:"required"`
	AccountID   string  `json:"account_id"`
	TeamID      string  `json:"team_id"`
	Fields      []Field `json:"fields" validate:"required"`
	Category    int     `json:"category"`
	State       int     `json:"state"`
	IsPublic    bool    `json:"is_public"`
	IsCore      bool    `json:"is_core"`
	IsShared    bool    `json:"is_shared"`
}

//State for the entity specifies the current state of the entity
const (
	StateTeamLevel    = 0
	StateAccountLevel = 1
)

//Category specifies the different type of entities
const (
	CategoryUnknown            = 0
	CategoryData               = 1
	CategoryAPI                = 2
	CategoryTimeSeries         = 3
	CategoryEmail              = 4
	CategoryUsers              = 5
	CategorySchedule           = 6
	CategoryDelay              = 7
	CategoryChildUnit          = 8
	CategoryTask               = 9
	CategoryEmailConfig        = 10
	CategoryFlow               = 11 // alais for actual flow
	CategoryNode               = 12 // alais for actual node
	CategoryNotes              = 13
	CategoryMeeting            = 14 // this is a type like task, email, notes
	CategoryCalendar           = 15 // this is integration
	CategoryEvent              = 16
	CategoryStream             = 17
	CategoryNotification       = 18
	CategoryVisitorsInvitation = 19
	CategorySubData            = 20
	CategoryTimeseries         = 21
	CategoryApprovals          = 22
)
