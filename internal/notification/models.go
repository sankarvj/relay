package notification

import "time"

type ClientRegister struct {
	AccountID   string    `db:"account_id" json:"account_id"`
	UserID      string    `db:"user_id" json:"user_id"`
	DeviceToken string    `db:"device_token" json:"device_token"`
	DeviceType  string    `db:"device_type" json:"device_type"`
	Status      int       `db:"status" json:"status"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   int64     `db:"updated_at" json:"updated_at"`
}
