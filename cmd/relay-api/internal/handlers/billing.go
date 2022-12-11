package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/payment"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
)

// Bill represents the Bill API method handler set.
type Bill struct {
	db            *sqlx.DB
	authenticator *auth.Authenticator
}

// Events handle the webhook calls from stripe
func (b *Bill) Events(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Billing.Events")
	defer span.End()

	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		return err
	}

	event := stripe.Event{}

	if err := json.Unmarshal(payload, &event); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Webhook error while parsing basic request. %v\n", err.Error())
		return err
	}

	// Replace this endpoint secret with your endpoint's unique secret
	// If you are testing with the CLI, find the secret by running 'stripe listen'
	// If you are using an endpoint defined with the API or dashboard, look in your webhook settings
	// at https://dashboard.stripe.com/webhooks
	stripePublishKey := util.ExpvarGet("stripe_publish_key")
	event, err = webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), stripePublishKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Webhook signature verification failed. %v\n", err)
		return err
	}

	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "customer.subscription.created":
		customerID := event.Data.Object["customer"].(string)
		trialStart, trialEnd := trailStartnEnd(event.Data.Object)
		status := event.Data.Object["status"].(string)
		product, quantity := product(event)

		err = updateAccPlan(ctx, customerID, status, product, quantity, trialStart, trialEnd, b.db)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  updating plan to the existing account failed. %v\n", err)
			return err
		}
	case "customer.subscription.deleted":
		customerID := event.Data.Object["customer"].(string)
		err = moveToFreePlan(ctx, customerID, b.db)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  updating plan to the existing account failed. %v\n", err)
			return err
		}

	case "customer.subscription.trial_will_end":
		log.Printf("trial_will_end event %+v", event)
	case "customer.subscription.updated":
		customerID := event.Data.Object["customer"].(string)
		trialStart, trialEnd := trailStartnEnd(event.Data.Object)
		status := event.Data.Object["status"].(string)
		product, quantity := product(event)

		err = updateAccPlan(ctx, customerID, status, product, quantity, trialStart, trialEnd, b.db)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  updating plan to the existing account failed. %v\n", err)
			return err
		}
	default:
		fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
	}

	return web.Respond(ctx, w, nil, http.StatusOK)
}

func (b *Bill) Portal(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Billing.Portal")
	defer span.End()

	accountID := params["account_id"]

	currentUser, err := user.RetrieveCurrentUser(ctx, accountID, b.db)
	if err != nil {
		err := errors.New("auth_cliams_missing_from_context") // value used in the UI dont change the string message.
		return web.NewRequestError(err, http.StatusForbidden)
	}
	stripeLiveKey := util.ExpvarGet("stripe_live_key")
	url, err := payment.CustomerPortal(ctx, params["account_id"], currentUser.ID, stripeLiveKey, b.db)
	if err != nil {
		err := errors.New("failed_to_handle_customer_portal_failed") // value used in the UI dont change the string message.
		return web.NewRequestError(err, http.StatusForbidden)
	}

	return web.Respond(ctx, w, url, http.StatusOK)
}

func product(event stripe.Event) (string, float64) {
	var quantity float64
	var product string
	items := event.Data.Object["items"].(map[string]interface{})
	if items != nil {
		data := items["data"].([]interface{})

		if len(data) > 0 {

			firstData := data[0].(map[string]interface{})
			if firstData != nil {
				quantity = firstData["quantity"].(float64)
				fprice := firstData["price"].(map[string]interface{})
				if fprice != nil {
					// price = fprice["id"].(string)
					product = fprice["product"].(string)
				}
			}
		}
	}
	return product, quantity
}

func updateAccPlan(ctx context.Context, stripeCusID, status, product string, quantity, trailStart, trailEnd float64, db *sqlx.DB) error {
	acc, err := account.RetrieveByStripeID(ctx, stripeCusID, db)
	if err != nil {
		return err
	}
	err = account.UpdateStripePlan(ctx, acc.ID, status, planForProduct(product), int(quantity), int64(trailStart), int64(trailEnd), db)
	if err != nil {
		return err
	}

	return nil
}

func moveToFreePlan(ctx context.Context, stripeCusID string, db *sqlx.DB) error {
	acc, err := account.RetrieveByStripeID(ctx, stripeCusID, db)
	if err != nil {
		return err
	}
	err = account.UpdateStripePlan(ctx, acc.ID, account.StatusActive, account.PlanFree, int(2), int64(-1), int64(-1), db)
	if err != nil {
		return err
	}
	return nil
}

func planForProduct(stripeProductID string) int {
	switch stripeProductID {
	case "prod_MmY5ZAE6ej1KGK":
		return account.PlanPro
	case "prod_MmY4md2J6XhoVr":
		return account.PlanStartup
	case "prod_MmY2qxRAjDQ99K":
		return account.PlanFree
	default:
		log.Println("unexpected error: unknown plan reached.... sending pro")
		return account.PlanPro
	}
}

func trailStartnEnd(obj map[string]interface{}) (float64, float64) {
	var trialStart, trialEnd float64
	if obj["trial_start"] != nil {
		trialStart = obj["trial_start"].(float64)
	}
	if obj["trial_end"] != nil {
		trialEnd = obj["trial_end"].(float64)
	}
	return trialStart, trialEnd
}
