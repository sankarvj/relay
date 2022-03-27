package email_test

import (
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/platform/integration/email"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
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
			toField := []interface{}{"vijayasankarmail@gmail.com"}
			subject := "Hi"
			body := "Hello"
			e := email.FallbackMail{Domain: "", ReplyTo: ""}
			_, err := e.SendMail("", "contact@wayplot.com", "", util.ConvertSliceTypeRev(toField), subject, body)
			if err != nil {
				t.Fatalf("\tShould be able to send the email during signup - %s", err)
			}

		}

	}
}
