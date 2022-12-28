package payment

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/billingportal/session"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/subscription"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

const (
	monthlyPro     = "price_1MCSkJHUBFGeRHv57UGsHmIF"
	yearlyPro      = "price_1MCSosHUBFGeRHv5mn6cu8IO"
	monthlyStartup = "price_1MCSrUHUBFGeRHv5YKAFFq3L"
	yearlyStartup  = "price_1MCSrUHUBFGeRHv5PBjvxjhq"
	monthlyFree    = "price_1MCSlqHUBFGeRHv5rrOnsPxu"
	yearlyFree     = "price_1MCSlqHUBFGeRHv5rrOnsPxu"
)

func InitStripe(ctx context.Context, accountID, userID string, stripeLiveKey string, db *sqlx.DB) error {
	user, err := user.RetrieveUser(ctx, db, accountID, userID)
	if err != nil {
		return err
	}
	cusID, err := addStripeCustomer(user.Email, stripeLiveKey)
	if err != nil {
		return err
	}

	trailStart, trialEnd, err := startTrail(cusID, stripeLiveKey)
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

func CustomerPortal(ctx context.Context, accountID, userID string, stripeLiveKey string, db *sqlx.DB) (string, error) {
	acc, err := account.Retrieve(ctx, db, accountID)
	if err != nil {
		return "", err
	}

	stripe.Key = stripeLiveKey
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

func AddStripCus(email, stripeLiveKey string) (string, error) {
	return addStripeCustomer(email, stripeLiveKey)
}

func addStripeCustomer(email, stripeLiveKey string) (string, error) {
	stripe.Key = stripeLiveKey
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
	}
	result, err := customer.New(params)
	if err != nil {
		return "", err
	}
	log.Printf("addStripeCustomer --- %+v", result)
	log.Println("addStripeCustomer err --->>>>", err)
	return result.ID, nil
}

func startTrail(cusID, stripeLiveKey string) (int64, int64, error) {
	stripe.Key = stripeLiveKey
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
