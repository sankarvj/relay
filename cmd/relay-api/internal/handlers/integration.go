package handlers

import (
	"io/ioutil"
	"log"
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
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
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
	ctx, span := trace.StartSpan(ctx, "internal.email.access")
	defer span.End()
	integrationID := params["integration_id"]
	var (
		authURL string
		err     error
	)
	switch integrationID {
	case integration.TypeGmail:
		authURL, err = integration.AskGmailAccessURL(ctx, g.authenticator.GoogleClientSecret)
	}
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, authURL, http.StatusOK)
}

func (g *Integration) SaveIntegration(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.email.saveIntegration")
	defer span.End()

	integrationID := params["integration_id"]
	accountID := params["account_id"]

	var code Code
	if err := web.Decode(r, &code); err != nil {
		return errors.Wrap(err, "cannot parse the code")
	}

	var (
		emailAddress string
		tokenJson    string
		err          error
	)
	switch integrationID {
	case integration.TypeGmail:
		tokenJson, err = integration.GetToken(g.authenticator.GoogleClientSecret, code.Code)
		if err != nil {
			return err
		}
		emailAddress, err = integration.WatchMessage(g.authenticator.GoogleClientSecret, tokenJson, g.publisher.Topic)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}

	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return err
	}

	_, err = entity.SaveEmailIntegration(ctx, accountID, currentUserID, integration.DomainGMail, tokenJson, emailAddress, g.db)
	if err != nil {
		return err
	}
	return web.Respond(ctx, w, "SUCCESS", http.StatusOK)
}

func (g *Integration) ReceiveEmail(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	log.Println("Received the push")
	return web.Respond(ctx, w, "SUCCESS", http.StatusOK)
}

func getConfig(oAuthFile string) (*oauth2.Config, error) {
	b, err := ioutil.ReadFile(oAuthFile)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read client secret file")
	}
	return google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
}

type Code struct {
	Code string `json:"code"`
}
