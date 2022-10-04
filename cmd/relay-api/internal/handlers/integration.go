package handlers

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/conversation"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration/calendar"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration/email"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
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
	sdb           *database.SecDB
	authenticator *auth.Authenticator
	publisher     *conversation.Publisher
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

func (i *Integration) SaveIntegration(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.integration.saveIntegration")
	defer span.End()

	var (
		emailAddress string
		tokenJson    string
		err          error
	)

	integrationID := params["integration_id"]
	accountID := params["account_id"]
	teamID := params["team_id"] //actually in the routes the teamID is not there
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
		tokenJson, err = integration.GetGoogleToken(i.authenticator.GoogleClientSecret, code.Code, integration.GmailScopes...)
		if err != nil {
			return errors.Wrapf(err, "Unable to get token from google")
		}

		g := email.Gmail{OAuthFile: i.authenticator.GoogleClientSecret, TokenJson: tokenJson}

		emailAddress, err = g.Watch(i.publisher.Topic)
		if err != nil {
			return err
		}

		discoveryID := emailAddress
		emailConfigEntityItem := entity.EmailConfigEntity{
			APIKey: tokenJson,
			Domain: integration.DomainGMail,
			Email:  emailAddress,
			Common: "false",
			Owner:  []string{currentUserID},
		}

		_, err = entity.SaveFixedEntityItem(ctx, accountID, teamID, currentUserID, entity.FixedEntityEmailConfig, "Gmail Config", discoveryID, integrationID, util.ConvertInterfaceToMap(emailConfigEntityItem), i.db)
		if err != nil {
			return errors.Wrapf(err, "Unable to create integration")
		}
	case integration.TypeGoogleCalendar:
		tokenJson, err = integration.GetGoogleToken(i.authenticator.GoogleClientSecret, code.Code, integration.GoogleCalendarScopes...)
		if err != nil {
			return errors.Wrapf(err, "Unable to get token from google")
		}

		calendarEntityItem := entity.CaldendarEntity{
			APIKey: tokenJson,
			ID:     "primary",
			Email:  emailAddress,
			Common: "false",
			Owner:  []string{currentUserID},
		}

		discoveryID := calendarEntityItem.ID
		c := calendar.Gcalendar{OAuthFile: i.authenticator.GoogleClientSecret, TokenJson: tokenJson}
		err = c.Watch(calendarEntityItem.ID, discoveryID)
		if err != nil {
			return errors.Wrapf(err, "Unable to watch event")
		}

		_, err = entity.SaveFixedEntityItem(ctx, accountID, teamID, currentUserID, entity.FixedEntityCalendar, "Google Calendar Config", discoveryID, integrationID, util.ConvertInterfaceToMap(calendarEntityItem), i.db)
		if err != nil {
			return errors.Wrapf(err, "Unable to create integration")
		}
	default:
		return web.Respond(ctx, w, "FAILURE", http.StatusNotImplemented)
	}
	return web.Respond(ctx, w, "SUCCESS", http.StatusOK)
}

type Code struct {
	Code string `json:"code"`
}
