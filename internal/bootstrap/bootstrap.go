package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/base"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/crm"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/csm"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

func BootstrapTeam(ctx context.Context, db *sqlx.DB, accountID, teamID, teamName string) error {
	const q = `INSERT INTO teams
		(team_id, account_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := db.ExecContext(
		ctx, q,
		teamID, accountID, teamName, "", time.Now().UTC(), time.Now().UTC().Unix(),
	)
	return err
}

func BootstrapOwnerEntity(ctx context.Context, currentUser *user.User, b *base.Base) error {
	fields, itemVals := forms.OwnerFields(currentUser.ID, *currentUser.Name, *currentUser.Avatar, currentUser.Email)
	// add entity - owners
	ue, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityOwner, "Owners", entity.CategoryUsers, entity.StateAccountLevel, fields)
	if err != nil {
		return err
	}
	// add owner item
	// pass the currentUserID as the itemID. Is it okay to do like that? seems like a anti pattern.
	_, err = b.ItemAdd(ctx, ue.ID, currentUser.ID, currentUser.ID, itemVals)
	if err != nil {
		return err
	}
	return nil
}

func BootstrapEmailConfigEntity(ctx context.Context, b *base.Base) error {
	coEntityID, coEmail, err := currentOwner(ctx, b.DB, b.AccountID, b.TeamID)
	if err != nil {
		return err
	}
	fields := forms.EmailConfigFields(coEntityID, coEmail)
	// add entity - email- configs
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityEmailConfig, "Email Integrations", entity.CategoryEmailConfig, entity.StateAccountLevel, fields)
	return err
}

func BootstrapCalendarEntity(ctx context.Context, b *base.Base) error {
	coEntityID, coEmail, err := currentOwner(ctx, b.DB, b.AccountID, b.TeamID)
	if err != nil {
		return err
	}
	fields := forms.CalendarFields(coEntityID, coEmail)
	// add entity - calendar
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityCalendar, "Calendar", entity.CategoryCalendar, entity.StateAccountLevel, fields)
	return err
}

func BootstrapNotificationEntity(ctx context.Context, b *base.Base) error {
	fields := forms.NotificationFields()
	// add entity - notifications
	_, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityNotification, "Notification", entity.CategoryNotification, entity.StateAccountLevel, fields)
	return err
}

func currentOwner(ctx context.Context, db *sqlx.DB, accountID, teamID string) (string, string, error) {
	ownerEntity, err := entity.RetrieveFixedEntity(ctx, db, accountID, teamID, entity.FixedEntityOwner)
	if err != nil {
		return "", "", err
	}
	ownerFields, err := ownerEntity.Fields()
	if err != nil {
		return "", "", err
	}
	return ownerEntity.ID, entity.NamedKeysMap(ownerFields)["email"], nil
}

// THE TEAM SPECIFIC BOOTS

func BootCRM(accountID string, db *sqlx.DB, rp *redis.Pool) error {
	fmt.Printf("\nBootstrap:CRM STARTED for the accountID %s\n", accountID)

	ctx := context.Background()
	teamID := uuid.New().String()
	err := BootstrapTeam(ctx, db, accountID, teamID, "CRM")
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:CRM `team` insertion failed")
	}
	fmt.Println("\t\t\tBootstrap:CRM `team` added")

	b := base.NewBase(accountID, teamID, base.UUIDHolder, db, rp)

	//boot
	fmt.Println("\t\t\tBootstrap:CRM `boot` functions started")
	err = crm.Boot(ctx, b)
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:CRM `boot` functions failed")
	}
	fmt.Println("Bootstrap:CRM `boot` functions completed successfully")

	//samples
	fmt.Println("Bootstrap:CRM `samples` functions started")
	err = crm.AddSamples(ctx, b)
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:CRM `samples` functions failed")
	}
	fmt.Println("\t\t\tBootstrap:CRM `samples` functions completed successfully")

	//event props
	fmt.Println("\t\t\tBootstrap:CRM `event props` functions started")
	err = crm.AddProps(ctx, b)
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:CRM `event props` functions failed")
	}
	fmt.Println("\t\t\tBootstrap:CRM `event props` functions completed successfully")

	//all done
	fmt.Printf("\nBootstrap:CRM ENDED successfully for the accountID: %s\n", accountID)

	return nil
}

func BootCSM(accountID string, db *sqlx.DB, rp *redis.Pool) error {
	fmt.Printf("\nBootstrap:CSM STARTED for the accountID %s\n", accountID)

	ctx := context.Background()
	teamID := uuid.New().String()
	err := BootstrapTeam(ctx, db, accountID, teamID, "CSM")
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:CSM `team` insertion failed")
	}
	fmt.Println("\t\t\tBootstrap:CSM `team` added")

	b := base.NewBase(accountID, teamID, base.UUIDHolder, db, rp)

	//boot
	fmt.Println("\t\t\tBootstrap:CSM `boot` functions started")
	err = csm.Boot(ctx, b)
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:CSM `boot` functions failed")
	}
	fmt.Println("\t\t\tBootstrap:CSM `boot` functions completed successfully")

	//all done
	fmt.Printf("\nBootstrap:CSM ENDED successfully for the accountID: %s\n", accountID)

	return nil
}
