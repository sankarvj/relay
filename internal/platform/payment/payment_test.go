package payment_test

import (
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/platform/payment"
)

func TestNewStripeCustomer(t *testing.T) {
	// _, teardown := tests.NewUnit(t)
	// defer teardown()

	//tests.SeedData(t, db)

	t.Log("Given the need to check payments using stripe")
	{
		t.Log("\tcreate a new customer in stripe")
		{
			_, err := payment.AddStripCus("jenny.rosen@example.com", "<Replace with live key>")
			if err != nil {
				t.Fatalf("\tshould create an stripe customer - %s", err)
			}
		}

	}
}

// func TestNewSubscription(t *testing.T) {
// 	// _, teardown := tests.NewUnit(t)
// 	// defer teardown()

// 	//tests.SeedData(t, db)

// 	t.Log("Given the need to check payments using stripe")
// 	{
// 		t.Log("\tcreate a new subscription in stripe")
// 		{
// 			err := payment.StartTrail(schema.SeedAccountID, "cus_MmfWLAU0O67vWY")
// 			if err != nil {
// 				t.Fatalf("\tShould create an stripe subscription for the given customer - %s", err)
// 			}
// 		}

// 	}
// }
