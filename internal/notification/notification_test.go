package notification_test

import (
	"fmt"
	"log"
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/notification"
	"gitlab.com/vjsideprojects/relay/internal/schema"
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
				AccountID:   schema.SeedAccountID,
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

func TestFirebase(t *testing.T) {
	const (
		credLoc     = "../../config/dev/relay-70013-firebase-adminsdk-cfun3-58caec85f0.json"
		clientToken = "f5xxp23sUPXwGKAn3NRVuB:APA91bFbGPG8wlJVwN21eBgzKAkIrM-IYcVEBzeMY2gLLkNnpgoW7VuoRxBqXpl8r5f5Xo94TtkHjwdO5iGIvt63k9A2Cy8YHiNPn99Vvr6rCo90Ay9FdkiczYlkbAjWaa3Y0AcIEKjx"
	)

	tp, _ := notification.NewTokenProvider(credLoc)
	tk, _ := tp.Token()
	log.Println("dd---- ", tk)

	t.Log(" Given the need to send firebase notification")
	{

		t.Log("\tSend firebase notification1")
		{
			data := make(map[string]string, 0)
			err := notification.FirebaseSend(data, "title", "message", clientToken, credLoc)
			if err != nil {
				t.Fatalf("\tShould be able to send the firebase notification - %s", err)
			}

		}

	}
}

func TestHostName(t *testing.T) {

	t.Log("Check hostname parse")
	{

		t.Log("\tshould parse the url properly")
		{

			hostname := notification.HostN("workbaseONE", "csp.workbaseone.com")
			if hostname != "workbaseone.workbaseone.com" {
				t.Fatalf("\tshould be able to parse URL %s", hostname)
			}

		}

	}
}
