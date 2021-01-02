package user

import (
	"time"

	"github.com/lib/pq"
)

// User represents someone with access to our system.
type User struct {
	ID           string         `db:"user_id" json:"id"`
	AccountIDs   pq.StringArray `db:"account_ids" json:"account_ids"`
	Name         *string        `db:"name" json:"name"`
	Avatar       *string        `db:"avatar" json:"avatar"`
	Email        string         `db:"email" json:"email"`
	Phone        *string        `db:"phone" json:"phone"`
	Verified     bool           `db:"verified" json:"verified"`
	Roles        pq.StringArray `db:"roles" json:"roles"`
	PasswordHash []byte         `db:"password_hash" json:"-"`
	Provider     *string        `db:"provider" json:"provider"`
	IssuedAt     *string        `db:"issued_at" json:"issued_at"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt    int64          `db:"updated_at" json:"updated_at"`
}

// NewUser contains information needed to create a new User.
type NewUser struct {
	AccountIDs      []string `json:"account_id" validate:"required"`
	Name            string   `json:"name" validate:"required"`
	Email           string   `json:"email" validate:"required"`
	Roles           []string `json:"roles" validate:"required"`
	Password        string   `json:"password" validate:"required"`
	PasswordConfirm string   `json:"password_confirm" validate:"eqfield=Password"`
}

type ViewModelUser struct {
	Name      string   `json:"name"`
	Avatar    string   `json:"avatar"`
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	Roles     []string `json:"roles"`
	CreatedAt string   `json:"created_at"`
}

// UpdateUser defines what information may be provided to modify an existing
// User. All fields are optional so clients can send just the fields they want
// changed. It uses pointer fields so we can differentiate between a field that
// was not provided and a field that was provided as explicitly blank. Normally
// we do not want to use pointers to basic types but we make exceptions around
// marshalling/unmarshalling.
type UpdateUser struct {
	AccountID       string   `json:"account_id"`
	Name            *string  `json:"name"`
	Email           *string  `json:"email"`
	Roles           []string `json:"roles"`
	Password        *string  `json:"password"`
	PasswordConfirm *string  `json:"password_confirm" validate:"omitempty,eqfield=Password"`
}
