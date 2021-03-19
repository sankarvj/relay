package bootstrap

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/schema"
)

func addEmails(ctx context.Context, db *sqlx.DB, accountID string, contactEntityID string, contactEntityKeyEmail, contactEntityKeyNPS string) (string, string, error) {
	//adding sandbox email-config item (this needs to be removed from here.)
	ei, err := entity.SaveEmailIntegration(ctx, accountID, schema.SeedUserID1, "sandbox3ab4868d173f4391805389718914b89c.mailgun.org", "9c2d8fbbab5c0ca5de49089c1e9777b3-7fba8a4e-b5d71e35", "vijayasankar.jothi@wayplot.com", db)
	if err != nil {
		return "", "", err
	}

	//adding sandbox email-template item (this needs to be removed from here.)
	to := fmt.Sprintf("{{%s.%s}}", contactEntityID, contactEntityKeyEmail)
	cc := "vijayasankarmobile@gmail.com"
	subject := fmt.Sprintf("This mail is sent you to tell that your NPS scrore is {{%s.%s}}. We are very proud of you!", contactEntityID, contactEntityKeyNPS)
	body := fmt.Sprintf("Hello {{%s.%s}}", contactEntityID, contactEntityKeyEmail)
	emg, err := entity.SaveEmailTemplate(ctx, accountID, ei.ID, schema.SeedUserID1, []string{to}, []string{cc}, []string{}, subject, body, db)
	if err != nil {
		return "", "", err
	}
	return ei.ID, emg.ID, nil
}
