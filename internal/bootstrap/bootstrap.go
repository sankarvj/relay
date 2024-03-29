package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/base"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/crm"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/csm"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/ctm"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/em"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/incident"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"

	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/team"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
)

func Bootstrap(ctx context.Context, db *sqlx.DB, sdb *database.SecDB, firebaseSDKPath string, accountID, accountName string, cuser *user.User) error {
	ctx, span := trace.StartSpan(ctx, "internal.account.Bootstrap")
	defer span.End()

	//Setting the accountID as the teamID for the base team of an account
	teamID := accountID
	memberID := uuid.New().String()

	//TODO: all bootsrapping should happen in a single transaction

	err := cuser.UpdateMemberID(ctx, memberID, db)
	if err != nil {
		return errors.Wrap(err, "user update with accounts failed")
	}

	err = BootstrapTeam(ctx, db, accountID, teamID, "base", "Base", "Base is the launchpad for custom apps")
	if err != nil {
		return errors.Wrap(err, "team bootstrap failed")
	}

	b := base.NewBase(accountID, teamID, cuser.ID, db, sdb, firebaseSDKPath)
	b.AccountName = accountName

	err = BootstrapOwnerEntity(ctx, memberID, cuser, b)
	if err != nil {
		return err
	}

	err = BootstrapNotificationEntity(ctx, b)
	if err != nil {
		return errors.Wrap(err, "notification bootstrap failed")
	}

	err = BootstrapEmailConfigEntity(ctx, b)
	if err != nil {
		return errors.Wrap(err, "email config bootstrap failed")
	}

	err = BootstrapCalendarEntity(ctx, b)
	if err != nil {
		return errors.Wrap(err, "calendar bootstrap failed")
	}

	err = BootstrapFlowAndNodeEntity(ctx, b)
	if err != nil {
		return errors.Wrap(err, "flow/node bootstrap failed")
	}

	err = BootstrapVisitorInviteEntity(ctx, b)
	if err != nil {
		return errors.Wrap(err, "visitor bootstrap failed")
	}

	err = BootstrapDelayEntity(ctx, b)
	if err != nil {
		return errors.Wrap(err, "delay bootstrap failed")
	}

	return nil
}

func BootstrapTeam(ctx context.Context, db *sqlx.DB, accountID, teamID, teamLookUp, teamName, description string) error {
	const q = `INSERT INTO teams
		(team_id, account_id, look_up, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := db.ExecContext(
		ctx, q,
		teamID, accountID, teamLookUp, teamName, description, time.Now().UTC(), time.Now().UTC().Unix(),
	)
	return err
}

func BootstrapOwnerEntity(ctx context.Context, memberID string, currentUser *user.User, b *base.Base) error {
	var err error
	fields, itemVals := forms.OwnerFields(b.TeamID, currentUser.ID, *currentUser.Name, *currentUser.Avatar, currentUser.Email, currentUser.Phone)
	// add entity - owners
	b.OwnerEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityOwner, "Owners", entity.CategoryUsers, entity.StateAccountLevel, false, false, true, fields)
	if err != nil {
		return err
	}
	//Adding currentUserID as the memberID for the first time
	_, err = b.ItemAdd(ctx, b.OwnerEntity.ID, memberID, currentUser.ID, itemVals, nil)
	if err != nil {
		return err
	}
	return nil
}

func BootstrapEmailConfigEntity(ctx context.Context, b *base.Base) error {

	coEntityID, coEmail, err := CurrentOwner(ctx, b.DB, b.AccountID, b.TeamID)
	if err != nil {
		return err
	}
	fields := forms.EmailConfigFields(coEntityID, coEmail)
	// add entity - email- configs
	b.EmailConfigEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityEmailConfig, "Email Integrations", entity.CategoryEmailConfig, entity.StateAccountLevel, false, false, true, fields)
	if err != nil {
		return err
	}

	baseMailUsername := "support"
	if b.AccountName != "" {
		baseMailUsername = b.AccountName
	}
	uniqueDiscoveryID := uuid.New().String()
	baseEmail := fmt.Sprintf("%s@%s.workbaseone.com", baseMailUsername, uniqueDiscoveryID)

	emailConfigInboxEntityItem := entity.EmailConfigEntity{
		AccountID: b.AccountID,
		TeamID:    b.TeamID,
		APIKey:    uniqueDiscoveryID,
		Domain:    integration.DomainBaseInbox,
		Email:     baseEmail,
		Common:    "true",
		Owner:     []string{b.UserID},
	}
	_, err = entity.SaveFixedEntityItem(ctx, b.AccountID, b.TeamID, schema.SeedSystemUserID, entity.FixedEntityEmailConfig, "Base Inbox Integration", uniqueDiscoveryID, integration.TypeBaseInbox, util.ConvertInterfaceToMap(emailConfigInboxEntityItem), b.DB)
	return err
}

func BootstrapCalendarEntity(ctx context.Context, b *base.Base) error {
	coEntityID, coEmail, err := CurrentOwner(ctx, b.DB, b.AccountID, b.TeamID)
	if err != nil {
		return err
	}
	fields := forms.CalendarFields(coEntityID, coEmail)
	// add entity - calendar
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityCalendar, "Calendar", entity.CategoryCalendar, entity.StateAccountLevel, false, false, true, fields)
	return err
}

func BootstrapNotificationEntity(ctx context.Context, b *base.Base) error {
	coEntityID, _, err := CurrentOwner(ctx, b.DB, b.AccountID, b.TeamID)
	if err != nil {
		return err
	}

	fields := forms.NotificationFields(coEntityID)
	// add entity - notifications
	_, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityNotification, "Notification", entity.CategoryNotification, entity.StateAccountLevel, false, false, true, fields)
	return err
}

func BootstrapFlowAndNodeEntity(ctx context.Context, b *base.Base) error {
	var err error
	// Flow wrapper entity added to facilitate other entities(deals) to reference the flows(pipeline) as the reference fields
	b.FlowEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityFlow, "Flow", entity.CategoryFlow, entity.StateAccountLevel, false, false, true, forms.FlowFields())
	if err != nil {
		return err
	}

	// Node wrapper entity added to facilitate other entities(deals) to reference the stages(pipeline stage) as the reference fields
	b.NodeEntity, err = b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityNode, "Node", entity.CategoryNode, entity.StateAccountLevel, false, false, true, forms.NodeFields())
	if err != nil {
		return err
	}
	fmt.Println("\tBOOT Flow & Node Wrapper Entities Created")

	return err
}

func BootstrapVisitorInviteEntity(ctx context.Context, b *base.Base) error {
	// add entity - task
	fields := forms.VisitorInvitationFields()
	_, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityVisitorInvite, "Visitors", entity.CategoryVisitorsInvitation, entity.StateAccountLevel, false, false, true, fields)
	if err != nil {
		return err
	}
	fmt.Println("\tBOOT VisitorInvitaiton Entity Created")
	return err
}

func BootstrapDelayEntity(ctx context.Context, b *base.Base) error {
	// add entity - delay
	_, err := b.EntityAdd(ctx, uuid.New().String(), entity.FixedEntityDelay, "Delay Timer", entity.CategoryDelay, entity.StateAccountLevel, false, false, true, forms.DelayFields())
	return err
}

func CurrentOwner(ctx context.Context, db *sqlx.DB, accountID, teamID string) (string, string, error) {
	ownerEntity, err := entity.RetrieveFixedEntity(ctx, db, accountID, teamID, entity.FixedEntityOwner)
	if err != nil {
		return "", "", err
	}
	ownerFields, err := ownerEntity.Fields()
	if err != nil {
		return "", "", err
	}
	return ownerEntity.ID, entity.NameKeyMap(ownerFields)["email"], nil
}

// THE TEAM SPECIFIC BOOTS

func BootCRM(accountID, userID string, db *sqlx.DB, sdb *database.SecDB, firebaseSDKPath string) error {
	fmt.Printf("\nBootstrap:CRM STARTED for the accountID %s\n", accountID)

	ctx := context.Background()
	teamID := uuid.New().String()
	crpTemplate := team.FindTeamTemplate(team.PredefinedTeamCRP)
	err := BootstrapTeam(ctx, db, accountID, teamID, crpTemplate.Key, crpTemplate.Name, crpTemplate.Description)
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:CRM `team` insertion failed")
	}
	fmt.Println("\t\t\tBootstrap:CRM `team` added")

	b := base.NewBase(accountID, teamID, userID, db, sdb, firebaseSDKPath)

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

	//workflows
	fmt.Println("Bootstrap:CRM `workflows` functions started")
	err = crm.AddWorkflows(ctx, b)
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:CRM `workflows` functions failed")
	}
	fmt.Println("\t\t\tBootstrap:CRM `workflows` functions completed successfully")

	//all done
	fmt.Printf("\nBootstrap:CRM ENDED successfully for the accountID: %s\n", accountID)

	return nil
}

func BootCSM(accountID, userID string, db *sqlx.DB, sdb *database.SecDB, firebaseSDKPath string) error {
	fmt.Printf("\nBootstrap:CSM STARTED for the accountID %s\n", accountID)

	ctx := context.Background()
	teamID := uuid.New().String()
	cspTemplate := team.FindTeamTemplate(team.PredefinedTeamCSP)
	err := BootstrapTeam(ctx, db, accountID, teamID, cspTemplate.Key, cspTemplate.Name, cspTemplate.Description)
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:CSM `team` insertion failed")
	}
	fmt.Println("\t\t\tBootstrap:CSM `team` added")

	b := base.NewBase(accountID, teamID, userID, db, sdb, firebaseSDKPath)

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

	//workflows
	fmt.Println("Bootstrap:CSM `workflows` functions started")
	err = csm.AddWorkflows(ctx, b)
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:CSM `workflows` functions failed")
	}
	fmt.Println("\t\t\tBootstrap:CSM `workflows` functions completed successfully")

	//all done
	fmt.Printf("\nBootstrap:CSP ENDED successfully for the accountID: %s\n", accountID)

	return nil
}

func BootEM(accountID, userID string, db *sqlx.DB, sdb *database.SecDB, firebaseSDKPath string) error {
	fmt.Printf("\nBootstrap:EM STARTED for the accountID %s\n", accountID)

	ctx := context.Background()
	teamID := uuid.New().String()
	empTemplate := team.FindTeamTemplate(team.PredefinedTeamEMP)
	err := BootstrapTeam(ctx, db, accountID, teamID, empTemplate.Key, empTemplate.Name, empTemplate.Description)
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:EM `team` insertion failed")
	}
	fmt.Println("\t\t\tBootstrap:EM `team` added")

	b := base.NewBase(accountID, teamID, userID, db, sdb, firebaseSDKPath)

	//boot
	fmt.Println("\t\t\tBootstrap:EM `boot` functions started")
	err = em.Boot(ctx, b)
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:EM `boot` functions failed")
	}
	fmt.Println("\t\t\tBootstrap:EM `boot` functions completed successfully")

	//all done
	fmt.Printf("\nBootstrap:EM ENDED successfully for the accountID: %s\n", accountID)

	return nil
}

func BootSupport(accountID, userID string, db *sqlx.DB, sdb *database.SecDB, firebaseSDKPath string) error {
	fmt.Printf("\nBootstrap:Support STARTED for the accountID %s\n", accountID)

	ctx := context.Background()
	teamID := uuid.New().String()
	supportTemplate := team.FindTeamTemplate(team.PredefinedTeamCSup)
	err := BootstrapTeam(ctx, db, accountID, teamID, supportTemplate.Key, supportTemplate.Name, supportTemplate.Description)
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:Support `team` insertion failed")
	}
	fmt.Println("\t\t\tBootstrap:Support `team` added")

	b := base.NewBase(accountID, teamID, userID, db, sdb, firebaseSDKPath)

	//boot
	fmt.Println("\t\t\tBootstrap:Support `boot` functions started")
	err = ctm.Boot(ctx, b)
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:Support `boot` functions failed")
	}
	fmt.Println("Bootstrap:Support `boot` functions completed successfully")

	//samples
	fmt.Println("Bootstrap:Support `samples` functions started")
	err = ctm.AddSamples(ctx, b)
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:Support `samples` functions failed")
	}
	fmt.Println("\t\t\tBootstrap:Support `samples` functions completed successfully")

	//all done
	fmt.Printf("\nBootstrap:Support ENDED successfully for the accountID: %s\n", accountID)

	return nil
}

func BootIncident(accountID, userID string, db *sqlx.DB, sdb *database.SecDB, firebaseSDKPath string) error {
	fmt.Printf("\nBootstrap:INCIDENT STARTED for the accountID %s\n", accountID)

	ctx := context.Background()
	teamID := uuid.New().String()
	incTemplate := team.FindTeamTemplate(team.PredefinedTeamINC)
	err := BootstrapTeam(ctx, db, accountID, teamID, incTemplate.Key, incTemplate.Name, incTemplate.Description)
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:INCIDENT `team` insertion failed")
	}
	fmt.Println("\t\t\tBootstrap:INCIDENT `team` added")

	b := base.NewBase(accountID, teamID, userID, db, sdb, firebaseSDKPath)

	//boot
	fmt.Println("\t\t\tBootstrap:INCIDENT `boot` functions started")
	err = incident.Boot(ctx, b)
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:INCIDENT `boot` functions failed")
	}
	fmt.Println("Bootstrap:INCIDENT `boot` functions completed successfully")

	//samples
	fmt.Println("Bootstrap:INCIDENT `samples` functions started")
	err = incident.AddSamples(ctx, b)
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:INCIDENT `samples` functions failed")
	}
	fmt.Println("\t\t\tBootstrap:INCIDENT `samples` functions completed successfully")

	//workflows
	fmt.Println("Bootstrap:INCIDENT `workflows` functions started")
	err = incident.AddWorkflows(ctx, b)
	if err != nil {
		return errors.Wrap(err, "\t\t\tBootstrap:INCIDENT `workflows` functions failed")
	}
	fmt.Println("\t\t\tBootstrap:INCIDENT `workflows` functions completed successfully")

	//all done
	fmt.Printf("\nBootstrap:INCIDENT ENDED successfully for the accountID: %s\n", accountID)

	return nil
}

// func BootPM(accountID, userID string, db *sqlx.DB, sdb *database.SecDB, firebaseSDKPath string) error {
// 	fmt.Printf("\nBootstrap:PM STARTED for the accountID %s\n", accountID)

// 	ctx := context.Background()
// 	teamID := uuid.New().String()
// 	err := BootstrapTeam(ctx, db, accountID, teamID, draft.TeamPM)
// 	if err != nil {
// 		return errors.Wrap(err, "\t\t\tBootstrap:PM `team` insertion failed")
// 	}
// 	fmt.Println("\t\t\tBootstrap:PM `team` added")

// 	b := base.NewBase(accountID, teamID, userID, db, sdb, firebaseSDKPath)

// 	//boot
// 	fmt.Println("\t\t\tBootstrap:PM `boot` functions started")
// 	err = pm.Boot(ctx, b)
// 	if err != nil {
// 		return errors.Wrap(err, "\t\t\tBootstrap:PM `boot` functions failed")
// 	}
// 	fmt.Println("\t\t\tBootstrap:PM `boot` functions completed successfully")

// 	//samples
// 	fmt.Println("Bootstrap:PM `samples` functions started")
// 	err = pm.AddSamples(ctx, b)
// 	if err != nil {
// 		return errors.Wrap(err, "\t\t\tBootstrap:PM `samples` functions failed")
// 	}
// 	fmt.Println("\t\t\tBootstrap:PM `samples` functions completed successfully")

// 	//all done
// 	fmt.Printf("\nBootstrap:PM ENDED successfully for the accountID: %s\n", accountID)

// 	return nil
// }
