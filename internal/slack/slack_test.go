package slack_test

import (
	"testing"

	"gitlab.com/vjsideprojects/relay/internal/slack"
)

const (
	slackIncomingWebHookUrl = "https://hooks.slack.com/services/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"
	slackBotToken           = "xoxb-3020548349907-3931927797877-FFl8wEJCmnpI4RpDFXMCmewM"
)

func TestSlackIncomingWebhook(t *testing.T) {
	// db, teardown := tests.NewUnit(t)
	// defer teardown()
	// tests.SeedData(t, db)

	t.Log(" Given the need to send a slack message using incoming webhook")
	{
		t.Log("\tWhen sending the simple text message")
		{
			err := slack.PostMessage(slackIncomingWebHookUrl, "Hello")
			if err != nil {
				t.Fatalf("\tShould be able to send the message successfully - %s", err)
			}
		}
	}
}

func TestSlackHomePublish(t *testing.T) {
	// db, teardown := tests.NewUnit(t)
	// defer teardown()
	// tests.SeedData(t, db)

	t.Log(" Given the need to send a home view for a user")
	{
		t.Log("\tWhen sending home view for the user")
		{
			err := slack.UpdateHomeView(slackBotToken, slack.MakeHomeView("U0305SJ3SQ7"))
			if err != nil {
				t.Fatalf("\tShould be able to send home view successfully - %s", err)
			}
		}
	}
}
