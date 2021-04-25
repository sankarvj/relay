package handlers

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"gitlab.com/vjsideprojects/relay/internal/platform/pubsub"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
	"golang.org/x/net/context"
)

var (
	// TokenNotFound is used when a specific token does not exist.
	TokenNotFound = errors.New("Token not found")
)

// Integration represents the data needed for the third party integration details.
type Integration struct {
	db            *sqlx.DB
	authenticator *auth.Authenticator
	publisher     *pubsub.Publisher
}

func (g *Integration) AccessIntegration(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "internal.integration.access")
	defer span.End()
	integrationID := params["integration_id"]
	var (
		accessURL string
		err       error
	)
	switch integrationID {
	case integration.TypeGmail:
		accessURL, err = integration.GetGoogleAccessURL(ctx, g.authenticator.GoogleClientSecret, integrationID, integration.GmailScopes...)
	case integration.TypeGoogleCalendar:
		accessURL, err = integration.GetGoogleAccessURL(ctx, g.authenticator.GoogleClientSecret, integrationID, integration.GoogleCalendarScopes...)
	default:
		return web.Respond(ctx, w, "FAILURE", http.StatusNotImplemented)
	}
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, accessURL, http.StatusOK)
}

func (g *Integration) SaveIntegration(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.integration.saveIntegration")
	defer span.End()

	var (
		emailAddress string
		tokenJson    string
		err          error
	)

	integrationID := params["integration_id"]
	accountID := params["account_id"]
	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return err
	}

	var code Code
	if err := web.Decode(r, &code); err != nil {
		return errors.Wrap(err, "cannot parse the code")
	}

	switch integrationID {
	case integration.TypeGmail:
		tokenJson, err = integration.GetGoogleToken(g.authenticator.GoogleClientSecret, code.Code, integration.GmailScopes...)
		if err != nil {
			return err
		}
		emailAddress, err = integration.WatchMessage(g.authenticator.GoogleClientSecret, tokenJson, g.publisher.Topic)
		if err != nil {
			return err
		}

		_, err = entity.SaveEmailIntegration(ctx, accountID, currentUserID, integration.DomainGMail, tokenJson, emailAddress, g.db)
		if err != nil {
			return err
		}
	case integration.TypeGoogleCalendar:
		tokenJson, err = integration.GetGoogleToken(g.authenticator.GoogleClientSecret, code.Code, integration.GoogleCalendarScopes...)
		if err != nil {
			return err
		}
		_, err = entity.SaveCalendarIntegration(ctx, accountID, currentUserID, integration.DomainGMail, tokenJson, emailAddress, g.db)
		if err != nil {
			return err
		}
	default:
		return web.Respond(ctx, w, "FAILURE", http.StatusNotImplemented)
	}

	return web.Respond(ctx, w, "SUCCESS", http.StatusOK)
}

type Code struct {
	Code string `json:"code"`
}
