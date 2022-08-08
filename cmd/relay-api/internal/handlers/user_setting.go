package handlers

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
)

func (u *User) RetriveUserSetting(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.UserSetting.Retrive")
	defer span.End()

	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return err
	}

	cus, err := user.UserSettingRetrieve(ctx, params["account_id"], currentUserID, u.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, createViewModelUS(cus), http.StatusOK)
}

// Update updates the specified user in the system.
func (u *User) UpdateUserSetting(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.UserSetting.UpdateUserSetting")
	defer span.End()

	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return err
	}

	var nus user.NewUserSetting
	if err := web.Decode(r, &nus); err != nil {
		return errors.Wrap(err, "")
	}

	nus.AccountID = params["account_id"]
	nus.UserID = currentUserID

	err = user.UpdateUserSettings(ctx, u.db, nus)
	if err != nil {
		return web.NewRequestError(err, http.StatusBadRequest)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func createViewModelUS(us user.UserSetting) user.ViewModelUserSetting {
	return user.ViewModelUserSetting{
		AccountID:           us.AccountID,
		UserID:              us.UserID,
		LayoutStyle:         us.LayoutStyle,
		SelectedTeam:        us.SelectedTeam,
		NotificationSetting: user.UnmarshalNotificationSettings(us.NotificationSetting),
	}
}
