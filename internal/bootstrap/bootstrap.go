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
	"go.opencensus.io/trace"
)

func Bootstrap(ctx context.Context, db *sqlx.DB, rp *redis.Pool, firebaseSDKPath string, accountID string, cuser *user.User) error {
	ctx, span := trace.StartSpan(ctx, "internal.account.Bootstrap")
	defer span.End()

	//Setting the accountID as the teamID for the base team of an account
	teamID := accountID

	//TODO: all bootsrapping should happen in a single transaction
	err := cuser.UpdateAccounts(ctx, db, map[string]interface{}{accountID: cuser.ID})
	if err != nil {
		return errors.Wrap(err, "account inserted but user update failed")
	}

	err = BootstrapTeam(ctx, db, accountID, teamID, "Base")
	if err != nil {
		return errors.Wrap(err, "account inserted but team bootstrap failed")
	}

	b := base.NewBase(accountID, teamID, cuser.ID, db, rp, firebaseSDKPath)

	err = BootstrapOwnerEntity(ctx, cuser, b)
	if err != nil {
		return err
	}

	err = BootstrapNotificationEntity(ctx, b)
	if err != nil {
		return errors.Wrap(err, "account inserted but notification bootstrap failed")
	}

	err = BootstrapEmailConfigEntity(ctx, b)
	if err != nil {
		return errors.Wrap(err, "account inserted but email config bootstrap failed")
	}

	err = BootstrapCalendarEntity(ctx, b)
	if err != nil {
		return errors.Wrap(err, "account inserted but calendar bootstrap failed")
	}

	err = BootstrapContactCompanyEntity(ctx, b)
	if err != nil {
		return errors.Wrap(err, "account inserted but contacts/companies bootstrap failed")
	}

	return nil
}

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

	fields, itemVals := forms.OwnerFields(b.TeamID, currentUser.ID, *currentUser.Name, *currentUser.Avatar, currentUser.Email)
	// add entity - owners
	ue, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityOwner, "Owners", entity.CategoryUsers, entity.StateAccountLevel, fields)
	if err != nil {
		return err
	}
	//Adding currentUserID as the memberID for the first time
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

func BootstrapContactCompanyEntity(ctx context.Context, b *base.Base) error {
	coEntityID, coEmail, err := currentOwner(ctx, b.DB, b.AccountID, b.TeamID)
	if err != nil {
		return err
	}

	// add entity - contacts
	conForms := forms.ContactFields(coEntityID, coEmail)
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityContacts, "Contacts", entity.CategoryData, entity.StateAccountLevel, conForms)
	if err != nil {
		return err
	}

	// add entity - companies
	comForms := forms.CompanyFields(coEntityID, coEmail)
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityCompanies, "Companies", entity.CategoryData, entity.StateAccountLevel, comForms)
	if err != nil {
		return err
	}
	return nil
}

func BootstrapNotificationEntity(ctx context.Context, b *base.Base) error {
	coEntityID, _, err := currentOwner(ctx, b.DB, b.AccountID, b.TeamID)
	if err != nil {
		return err
	}

	fields := forms.NotificationFields(coEntityID)
	// add entity - notifications
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityNotification, "Notification", entity.CategoryNotification, entity.StateAccountLevel, fields)
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

func BootCRM(accountID, userID string, db *sqlx.DB, rp *redis.Pool, firebaseSDKPath string) error {
	fmt.Printf("\nBootstrap:CRM STARTED for the accountID %s\n", accountID)

	ctx := context.Background()
	teamID := uuid.New().String()
	err := BootstrapTeam(ctx, db, accountID, teamID, "CRM")
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:CRM `team` insertion failed")
	}
	fmt.Println("\t\t\tBootstrap:CRM `team` added")

	b := base.NewBase(accountID, teamID, userID, db, rp, firebaseSDKPath)

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

func BootCSM(accountID, userID string, db *sqlx.DB, rp *redis.Pool, firebaseSDKPath string) error {
	fmt.Printf("\nBootstrap:CSM STARTED for the accountID %s\n", accountID)

	ctx := context.Background()
	teamID := uuid.New().String()
	err := BootstrapTeam(ctx, db, accountID, teamID, "CSM")
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:CSM `team` insertion failed")
	}
	fmt.Println("\t\t\tBootstrap:CSM `team` added")

	b := base.NewBase(accountID, teamID, userID, db, rp, firebaseSDKPath)

	//boot
	fmt.Println("\t\t\tBootstrap:CSM `boot` functions started")
	err = csm.Boot(ctx, b)
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:CSM `boot` functions failed")
	}
	fmt.Println("\t\t\tBootstrap:CSM `boot` functions completed successfully")

	//samples
	fmt.Println("Bootstrap:CSM `samples` functions started")
	err = csm.AddSamples(ctx, b)
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:CSM `samples` functions failed")
	}
	fmt.Println("\t\t\tBootstrap:CSM `samples` functions completed successfully")

	//all done
	fmt.Printf("\nBootstrap:CSM ENDED successfully for the accountID: %s\n", accountID)

	return nil
}
