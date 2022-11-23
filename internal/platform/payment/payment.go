package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stripe/stripe-go/v73"
	"github.com/stripe/stripe-go/v73/BillingPortal/session"
	"github.com/stripe/stripe-go/v73/customer"
	"github.com/stripe/stripe-go/v73/subscription"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

const (
	stripeTestKey  = "sk_test_51M0BSXHUBFGeRHv5Qalelfhv8NO1kdnM0FgGd37iG74b2HNQfRLSolOgcvuFjvkfRP4KYTmZwztk5qMCmN245IDW00IUDFBOmp"
	EndPointSecret = "whsec_41d7022cc154e767fe96054ac413c1cde21b2d9c23b4c7743f20315901f247cc"
)

const (
	monthlyPro     = "price_1M2yysHUBFGeRHv5ny0GpDcg"
	yearlyPro      = "price_1M33FmHUBFGeRHv5CacjHjmY"
	monthlyStartup = "price_1M2yy2HUBFGeRHv5Egzwph1J"
	yearlyStartup  = "price_1M33I1HUBFGeRHv5JbtHIUIC"
	monthlyFree    = "price_1M2yw0HUBFGeRHv5pQzB0gPT"
	yearlyFree     = "price_1M2yw0HUBFGeRHv5j65gw5i4"
)

func InitStripe(ctx context.Context, accountID, userID string, db *sqlx.DB) error {
	user, err := user.RetrieveUser(ctx, db, accountID, userID)
	if err != nil {
		return err
	}
	cusID, err := addStripeCustomer(user.Email)
	if err != nil {
		return err
	}

	trailStart, trialEnd, err := startTrail(cusID)
	if err != nil {
		return err
	}

	//update the account with e-mail and cusID and set the plan as pro for trail
	err = account.UpdateStripeCustomer(ctx, accountID, cusID, user.Email, account.StatusTrial, account.PlanPro, trailStart, trialEnd, db)
	if err != nil {
		return err
	}

	return nil
}

func CustomerPortal(ctx context.Context, accountID, userID string, db *sqlx.DB) (string, error) {
	acc, err := account.Retrieve(ctx, db, accountID)
	if err != nil {
		return "", err
	}

	stripe.Key = stripeTestKey
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(*acc.CustomerID),
		ReturnURL: stripe.String(billingLink(acc.ID, acc.Domain)),
	}
	s, err := session.New(params)
	if err != nil {
		return "", err
	}

	return s.URL, nil
}

func addStripeCustomer(email string) (string, error) {
	stripe.Key = stripeTestKey
	params := &stripe.CustomerParams{
		Email:         stripe.String(email),
		PaymentMethod: stripe.String("pm_card_visa"),
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String("pm_card_visa"),
		},
	}
	result, err := customer.New(params)
	if err != nil {
		return "", err
	}
	return result.ID, nil
}

func startTrail(cusID string) (int64, int64, error) {
	stripe.Key = stripeTestKey
	items := []*stripe.SubscriptionItemsParams{
		{
			Price: stripe.String(monthlyPro),
		},
	}
	params := &stripe.SubscriptionParams{
		Customer:        stripe.String(cusID),
		Items:           items,
		TrialPeriodDays: stripe.Int64(14),
	}
	sub, err := subscription.New(params)
	if err != nil {
		return util.GetMilliSeconds(time.Now()), util.GetMilliSeconds(time.Now()), err
	}

	return sub.TrialStart, sub.TrialEnd, err
}

func billingLink(accountID, accountDomain string) string {
	billLink := fmt.Sprintf("https://%s/v1/accounts/%s/billing", accountDomain, accountID)
	return billLink
}
