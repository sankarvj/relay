package bootstrap

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/schema"
)

func addEmails(ctx context.Context, db *sqlx.DB, accountID string, contactEntityID string, contactEntityKeyEmail, contactEntityKeyNPS string) error {
	emailConfigEntityItem := entity.EmailConfigEntity{
		APIKey: "9c2d8fbbab5c0ca5de49089c1e9777b3-7fba8a4e-b5d71e35",
		Domain: integration.DomainMailGun,
		Email:  "vijayasankar.jothi@wayplot.com",
		Common: "false",
		Owner:  []string{schema.SeedUserID1},
	}
	err := entity.SaveFixedEntityItem(ctx, accountID, schema.SeedUserID1, entity.FixedEntityEmailConfig, "Mail Gun Integration", "vijayasankar.jothi@wayplot.com", integration.TypeMailGun, util.ConvertInterfaceToMap(emailConfigEntityItem), db)
	if err != nil {
		return err
	}

	emailEntityItem := entity.EmailEntity{
		From:    []string{},
		To:      []string{fmt.Sprintf("{{%s.%s}}", contactEntityID, contactEntityKeyEmail)},
		Cc:      []string{"vijayasankarmobile@gmail.com"},
		Bcc:     []string{""},
		Subject: fmt.Sprintf("This mail is sent you to tell that your NPS scrore is {{%s.%s}}. We are very proud of you!", contactEntityID, contactEntityKeyNPS),
		Body:    fmt.Sprintf("Hello {{%s.%s}}", contactEntityID, contactEntityKeyEmail),
	}

	err = entity.SaveFixedEntityItem(ctx, accountID, schema.SeedUserID1, entity.FixedEntityEmails, "Cult Mail Template", "", "", util.ConvertInterfaceToMap(emailEntityItem), db)
	if err != nil {
		return err
	}
	return nil
}
