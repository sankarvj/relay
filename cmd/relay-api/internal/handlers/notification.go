package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/notification"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
)

type Notification struct {
	db            *sqlx.DB
	sdb           *database.SecDB
	authenticator *auth.Authenticator
}

func (n *Notification) Register(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Notification.Register")
	defer span.End()

	accountID := params["account_id"]

	var cr notification.ViewModelClientRegister
	if err := web.Decode(r, &cr); err != nil {
		return errors.Wrap(err, "")
	}

	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return errors.Wrapf(err, "auth claims missing from context")
	}

	if cr.DeviceToken != "" {
		_, err = notification.RetrieveClient(ctx, accountID, currentUserID, cr.DeviceToken, n.db)
		if err == notification.ErrNotFound {
			_, err = notification.CreateClient(ctx, n.db, accountID, currentUserID, cr, time.Now())
			if err != nil {
				return errors.Wrapf(err, "failure in saving the client token")
			}
		}
	}
	if cr.OldToken != "" {
		err = notification.DeleteClient(ctx, accountID, currentUserID, cr.OldToken, n.db)
		if err != nil {
			return errors.Wrapf(err, "failure in saving the client token")
		}
	}

	return web.Respond(ctx, w, true, http.StatusOK)
}

func (n *Notification) Clear(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Notification.Clear")
	defer span.End()

	accountID, entityID, itemID := takeAEI(ctx, params, n.db)

	currentUser, err := user.RetrieveCurrentUser(ctx, n.db)
	if err != nil {
		return err
	}

	e, err := entity.Retrieve(ctx, accountID, entityID, n.db)
	if err != nil {
		return err
	}

	existingItem, err := item.Retrieve(ctx, entityID, itemID, n.db)
	if err != nil {
		return errors.Wrapf(err, "get item when the notification clear")
	}
	updatedFields := existingItem.Fields()

	whoMap := e.WhoFields()

	if memberID, ok := currentUser.AccountsB()[accountID]; ok {
		followers := updatedFields[whoMap[entity.WhoFollower]]
		if followers != nil {
			followers := followers.([]interface{})
			for i, fID := range followers {
				if fID == memberID {
					updatedFields[whoMap[entity.WhoFollower]] = removeIndex(followers, i)
					break
				}
			}
		}
		assigneeKey := whoMap[entity.WhoAssignee]
		assignees := updatedFields[whoMap[entity.WhoFollower]]
		if assignees != nil {
			assignees := assignees.([]interface{})
			for i, fID := range assignees {
				if fID == memberID {
					updatedFields[assigneeKey] = removeIndex(assignees, i)
					break
				}
			}
		}
	}

	it, err := item.UpdateFields(ctx, n.db, entityID, itemID, updatedFields)
	if err != nil {
		return errors.Wrapf(err, "Notification clear")
	}
	//stream
	go job.NewJob(n.db, n.sdb, n.authenticator.FireBaseAdminSDK).Stream(stream.NewUpdateItemMessage(ctx, n.db, accountID, currentUser.ID, entityID, it.ID, it.Fields(), existingItem.Fields()))

	return web.Respond(ctx, w, createViewModelItem(it), http.StatusOK)
}

func removeIndex(s []interface{}, index int) []interface{} {
	return append(s[:index], s[index+1:]...)
}
