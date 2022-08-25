package user

import (
	"time"

	"github.com/lib/pq"
)

const (
	UUID_ENGINE_USER = "10000000-1000-1000-1000-100000000000" //user calling from engine. executar data..
	UUID_SYSTEM_USER = "00000000-0000-0000-0000-000000000000" //system user
)

// User represents someone with access to our system.
type User struct {
	ID           string         `db:"user_id" json:"id"`
	Accounts     *string        `db:"accounts" json:"accounts"`
	Name         *string        `db:"name" json:"name"`
	Avatar       *string        `db:"avatar" json:"avatar"`
	Email        string         `db:"email" json:"email"`
	Phone        *string        `db:"phone" json:"phone"`
	Verified     bool           `db:"verified" json:"verified"`
	Roles        pq.StringArray `db:"roles" json:"roles"`
	PasswordHash []byte         `db:"password_hash" json:"-"`
	Provider     *string        `db:"provider" json:"provider"`
	IssuedAt     time.Time      `db:"issued_at" json:"issued_at"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt    int64          `db:"updated_at" json:"updated_at"`
}

type UserSetting struct {
	AccountID           string `db:"account_id" json:"account_id"`
	UserID              string `db:"user_id" json:"user_id"`
	LayoutStyle         string `db:"layout_style" json:"layout_style"`
	SelectedTeam        string `db:"selected_team" json:"selected_team"`
	NotificationSetting string `db:"notification_setting" json:"notification_setting"`
}

type NewUserSetting struct {
	AccountID           string            `json:"account_id"`
	UserID              string            `json:"user_id"`
	LayoutStyle         string            `json:"layout_style"`
	SelectedTeam        string            `json:"selected_team"`
	NotificationSetting map[string]string `json:"notification_setting"`
}

type ViewModelUserSetting struct {
	AccountID           string            `json:"account_id"`
	UserID              string            `json:"user_id"`
	LayoutStyle         string            `json:"layout_style"`
	SelectedTeam        string            `json:"selected_team"`
	NotificationSetting map[string]string `json:"notification_setting"`
}

// NewUser contains information needed to create a new User.
type NewUser struct {
	Accounts        map[string]interface{} `json:"accounts" validate:"required"`
	Name            string                 `json:"name" validate:"required"`
	Email           string                 `json:"email" validate:"required"`
	Avatar          *string                `json:"avatar"`
	Phone           *string                `json:"phone"`
	Provider        *string                `json:"provider"`
	Verified        bool                   `json:"verified"`
	Roles           []string               `json:"roles" validate:"required"`
	Password        string                 `json:"password" validate:"required"`
	PasswordConfirm string                 `json:"password_confirm" validate:"eqfield=Password"`
}

type ViewModelUser struct {
	Name   string   `json:"name"`
	Avatar string   `json:"avatar"`
	Email  string   `json:"email"`
	Phone  string   `json:"phone"`
	Roles  []string `json:"roles"`
}

// UpdateUser defines what information may be provided to modify an existing
// User. All fields are optional so clients can send just the fields they want
// changed. It uses pointer fields so we can differentiate between a field that
// was not provided and a field that was provided as explicitly blank. Normally
// we do not want to use pointers to basic types but we make exceptions around
// marshalling/unmarshalling.
type UpdateUser struct {
	Name            *string  `json:"name"`
	Email           *string  `json:"email"`
	Avatar          *string  `json:"avatar"`
	Phone           *string  `json:"phone"`
	Roles           []string `json:"roles"`
	Password        *string  `json:"password"`
	PasswordConfirm *string  `json:"password_confirm" validate:"omitempty,eqfield=Password"`
}
