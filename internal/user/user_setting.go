package user

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

func UserSettingRetrieve(ctx context.Context, accountID, userID string, db *sqlx.DB) (UserSetting, error) {
	ctx, span := trace.StartSpan(ctx, "internal.UserSetting.Retrieve")
	defer span.End()

	var us UserSetting
	const q = `SELECT * FROM user_settings WHERE account_id = $1 AND user_id = $2`
	if err := db.GetContext(ctx, &us, q, accountID, userID); err != nil {
		if err == sql.ErrNoRows {
			notificationSettingBytes, _ := json.Marshal(defaultNotificationSettings())
			return UserSetting{
				AccountID:           accountID,
				UserID:              userID,
				LayoutStyle:         "menu",
				SelectedTeam:        "",
				NotificationSetting: string(notificationSettingBytes),
			}, nil
		}

		return UserSetting{}, errors.Wrapf(err, "selecting user settings %q", userID)
	}

	return us, nil
}

func AddUserSetting(ctx context.Context, db *sqlx.DB, nus NewUserSetting) (UserSetting, error) {
	ctx, span := trace.StartSpan(ctx, "internal.UserSetting.AddUserSetting")
	defer span.End()

	notificationSettingBytes, err := json.Marshal(nus.NotificationSetting)
	if err != nil {
		return UserSetting{}, errors.Wrap(err, "encode notification settings to bytes")
	}

	us := UserSetting{
		AccountID:           nus.AccountID,
		UserID:              nus.UserID,
		LayoutStyle:         nus.LayoutStyle,
		SelectedTeam:        nus.SelectedTeam,
		NotificationSetting: string(notificationSettingBytes),
	}

	const q = `INSERT INTO user_settings
		(account_id, user_id, layout_style, selected_team, notification_setting)
		VALUES ($1, $2, $3, $4, $5)`

	_, err = db.ExecContext(
		ctx, q,
		us.AccountID, us.UserID, us.LayoutStyle, us.SelectedTeam, us.NotificationSetting,
	)
	if err != nil {
		return UserSetting{}, errors.Wrap(err, "inserting user setting")
	}

	return us, nil
}

func UpdateUserSettings(ctx context.Context, db *sqlx.DB, nus NewUserSetting) error {
	ctx, span := trace.StartSpan(ctx, "internal.UserSetting.UpdateUserSettings")
	defer span.End()

	// if not exist add it
	var us UserSetting
	const rq = `SELECT * FROM user_settings WHERE account_id = $1 AND user_id = $2`
	if err := db.GetContext(ctx, &us, rq, nus.AccountID, nus.UserID); err != nil {
		if err == sql.ErrNoRows {
			_, err = AddUserSetting(ctx, db, nus)
			if err != nil {
				return errors.Wrap(err, "encode notification settings to bytes")
			}
			return nil
		}
	}

	notificationSettingBytes, err := json.Marshal(nus.NotificationSetting)
	if err != nil {
		return errors.Wrap(err, "encode notification settings to bytes")
	}

	const q = `UPDATE user_settings SET
		"layout_style" = $3,
		"selected_team" = $4,
		"notification_setting" = $5 
		WHERE account_id = $1 AND user_id = $2`
	_, err = db.ExecContext(ctx, q, nus.AccountID, nus.UserID,
		nus.LayoutStyle, nus.SelectedTeam, string(notificationSettingBytes),
	)

	return err
}

func UnmarshalNotificationSettings(notificationSettingsB string) map[string]string {
	var notificationSettingsMap map[string]string
	if err := json.Unmarshal([]byte(notificationSettingsB), &notificationSettingsMap); err != nil {
		return notificationSettingsMap
	}
	return notificationSettingsMap
}

func defaultNotificationSettings() map[string]string {
	nSettings := make(map[string]string, 0)
	nSettings["updated"] = "true"
	nSettings["created"] = "true"
	nSettings["assigned"] = "true"
	return nSettings
}
