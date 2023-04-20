package voip_test

import (
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/platform/voip"
	"gitlab.com/vjsideprojects/relay/internal/schema"
)

func TestMakeCall(t *testing.T) {
	t.Log("\twhen calling a twilio voice.")
	{
		// db, teardown := tests.NewUnit(t)
		// defer teardown()
		twilioSID := "AC7a07a47f63e7a1266ae6ed7b7a9e95ba"
		twilioToken := "17389439db78673a488bc6b64a00feca"

		err := voip.MakeCall(twilioSID, twilioToken, schema.SeedAccountID, "", "69d2af4e-61a5-4239-988c-d5e1bd7bc4e7", "49a53e08-033a-40a2-912f-171a940e66e6", "+14083485853")
		if err != nil {
			t.Fatalf("\tShould be able to call that number - %s", err)
		}
	}
}
