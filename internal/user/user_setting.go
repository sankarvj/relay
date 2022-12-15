package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

const (
	NSUpdated           = "updated"
	NSCreated           = "created"
	NSAssigned          = "assigned"
	NSEmailSubscription = "email_subscription"
)

func RetrieveUserSetting(ctx context.Context, db *sqlx.DB, accountID, userID string) (*NotificationUserSetting, error) {
	var u NotificationUserSetting
	const q = `SELECT u.user_id, u.member_id, u.name, u.avatar, u.email, us.notification_setting FROM users as u join user_settings as us on u.user_id = us.user_id WHERE u.user_id = $1 AND u.account_id = $2`
	if err := db.GetContext(ctx, &u, q, userID, accountID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, errors.Wrapf(err, "selecting current user %q", userID)
	}

	return &u, nil
}

func BulkRetrieveUserSetting(ctx context.Context, db *sqlx.DB, accountID string, ids []interface{}) ([]NotificationUserSetting, error) {
	users := []NotificationUserSetting{}
	const q = `SELECT u.user_id, u.member_id, u.name, u.avatar, u.email, us.notification_setting FROM users as u join user_settings as us on u.user_id = us.user_id WHERE u.account_id = $1 AND u.user_id = any($2)`

	if err := db.SelectContext(ctx, &users, q, accountID, pq.Array(ids)); err != nil {
		return users, errors.Wrap(err, "selecting items for a list of item ids")
	}

	return users, nil
}

func UserSettingRetrieve(ctx context.Context, accountID, userID string, db *sqlx.DB) (UserSetting, error) {
	ctx, span := trace.StartSpan(ctx, "internal.UserSetting.Retrieve")
	defer span.End()

	var us UserSetting
	const q = `SELECT * FROM user_settings WHERE account_id = $1 AND user_id = $2`
	if err := db.GetContext(ctx, &us, q, accountID, userID); err != nil {
		if err == sql.ErrNoRows {
			notificationSettingBytes, _ := json.Marshal(defaultNotificationSettings(true))
			return UserSetting{
				AccountID:           accountID,
				UserID:              userID,
				SelectedTeam:        "",
				SelectedEntity:      "",
				SelectedView:        "",
				SelectedOrder:       "",
				SelectedTheme:       "light-theme-1",
				Metab:               string(MarshalMeta(map[string]string{})),
				LayoutStyle:         "menu",
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
		SelectedTeam:        nus.SelectedTeam,
		SelectedEntity:      nus.SelectedEntity,
		SelectedView:        nus.SelectedView,
		SelectedOrder:       nus.SelectedOrder,
		SelectedTheme:       nus.SelectedTheme,
		LayoutStyle:         nus.LayoutStyle,
		Metab:               string(MarshalMeta(nus.Meta)),
		NotificationSetting: string(notificationSettingBytes),
	}

	const q = `INSERT INTO user_settings
		(account_id, user_id, selected_team, selected_entity, selected_view, selected_order, selected_theme, layout_style, metab, notification_setting)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err = db.ExecContext(
		ctx, q,
		us.AccountID, us.UserID, us.SelectedTeam, us.SelectedEntity, us.SelectedView, us.SelectedOrder, us.SelectedTheme, us.LayoutStyle, us.Metab, us.NotificationSetting,
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
		"selected_team" = $3,
		"selected_entity" = $4,
		"selected_view" = $5,
		"selected_order" = $6,
		"selected_theme" = $7,
		"layout_style" = $8,
		"metab" = $9,
		"notification_setting" = $10 
		WHERE account_id = $1 AND user_id = $2`
	_, err = db.ExecContext(ctx, q, nus.AccountID, nus.UserID, nus.SelectedTeam, nus.SelectedEntity, nus.SelectedView, nus.SelectedOrder, nus.SelectedTheme,
		nus.LayoutStyle, string(MarshalMeta(nus.Meta)), string(notificationSettingBytes),
	)

	return err
}

func UpdateEmailSubscription(ctx context.Context, db *sqlx.DB, accountID, userID string, emailSubscription bool) error {
	ctx, span := trace.StartSpan(ctx, "internal.UserSetting.UpdateUserSettings")
	defer span.End()

	// if not exist add it
	var us UserSetting
	const rq = `SELECT * FROM user_settings WHERE account_id = $1 AND user_id = $2`
	if err := db.GetContext(ctx, &us, rq, accountID, userID); err != nil {
		if err == sql.ErrNoRows {
			nus := NewUserSetting{
				AccountID:           accountID,
				UserID:              userID,
				SelectedTeam:        "",
				SelectedEntity:      "",
				SelectedView:        "",
				SelectedOrder:       "",
				SelectedTheme:       "light-theme-1",
				Meta:                map[string]string{},
				LayoutStyle:         "menu",
				NotificationSetting: defaultNotificationSettings(emailSubscription),
			}
			_, err = AddUserSetting(ctx, db, nus)
			if err != nil {
				return errors.Wrap(err, "encode notification settings to bytes")
			}
			return nil
		}
	}

	nsMap := UnmarshalNotificationSettings(us.NotificationSetting)
	nsMap["email_subscription"] = strconv.FormatBool(emailSubscription)
	notificationSettingBytes, err := json.Marshal(nsMap)
	if err != nil {
		return errors.Wrap(err, "encode notification settings to bytes")
	}

	const q = `UPDATE user_settings SET
		"notification_setting" = $3 
		WHERE account_id = $1 AND user_id = $2`
	_, err = db.ExecContext(ctx, q, accountID, userID,
		string(notificationSettingBytes),
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

func defaultNotificationSettings(emailSubscription bool) map[string]string {
	nSettings := make(map[string]string, 0)
	nSettings[NSUpdated] = "true"
	nSettings[NSCreated] = "true"
	nSettings[NSAssigned] = "true"
	nSettings[NSEmailSubscription] = strconv.FormatBool(emailSubscription)
	return nSettings
}

func MarshalMeta(meta map[string]string) []byte {
	json, err := json.Marshal(meta)
	if err != nil {
		log.Println("***> unexpected/unhandled error in internal.usersettings. when marshaling meta. error:", err)
	}

	return json
}

func UnmarshalMeta(metaB string) map[string]string {
	var meta map[string]string
	if err := json.Unmarshal([]byte(metaB), &meta); err != nil {
		return meta
	}
	return meta
}
