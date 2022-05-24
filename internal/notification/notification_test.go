package notification_test

import (
	"fmt"
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/notification"
	"gitlab.com/vjsideprojects/relay/internal/tests"
)

func TestEmail(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	defer teardown()
	tests.SeedData(t, db)

	t.Log(" Given the need to send emails")
	{

		t.Log("\tSend email during the signup")
		{

			emailNotif := notification.EmailNotification{
				To:          []interface{}{"vijayasankarmail@gmail.com"},
				Subject:     fmt.Sprintf("%s is ready", "Acme"),
				Name:        "Tester",
				AccountName: "Acme",
				MagicLink:   "https://workbaseone.com/home",
			}
			err := emailNotif.Send(tests.Context(), notification.TypeWelcome, db)
			if err != nil {
				t.Fatalf("\tShould be able to send the email during signup - %s", err)
			}

		}

	}
}
