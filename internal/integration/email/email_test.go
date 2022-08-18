package email_test

import (
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/platform/integration/email"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/tests"
)

func TestSESEmail(t *testing.T) {
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
			e := email.SESMail{Domain: "", ReplyTo: ""}
			_, err := e.SendMail("", "support@workbaseone.com", "", util.ConvertSliceTypeRev(toField), subject, body)
			if err != nil {
				t.Fatalf("\tShould be able to send the email during signup - %s", err)
			}

		}

	}
}

func TestSMTPEmail(t *testing.T) {
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
			e := email.SMTPMail{Domain: "", ReplyTo: ""}
			_, err := e.SendMail("", "support@workbaseone.com", "", util.ConvertSliceTypeRev(toField), subject, body)
			if err != nil {
				t.Fatalf("\tShould be able to send the email during signup - %s", err)
			}

		}

	}
}
