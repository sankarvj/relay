package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/base"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/crm"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap/forms"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/team"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
)

// Team represents the Team API method handler set.
type Team struct {
	db            *sqlx.DB
	rPool         *redis.Pool
	authenticator *auth.Authenticator
}

// List returns all the existing accounts associated with logged-in user
func (t *Team) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Team.List")
	defer span.End()

	teams, err := team.List(ctx, params["account_id"], t.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, teams, http.StatusOK)
}

func (t *Team) AddTemplates(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Team.TemplatesAdd")
	defer span.End()

	var nt team.Template
	if err := web.Decode(r, &nt); err != nil {
		return errors.Wrap(err, "")
	}
	currentUserID, _ := user.RetrieveCurrentUserID(ctx)

	err := t.createCustomTemplates(ctx, params["account_id"], currentUserID, nt.Key)
	if err != nil {
		err := errors.New("problem_adding_custom_templates") // value used in the UI dont change the string message.
		return web.NewRequestError(err, http.StatusConflict)
	}

	return web.Respond(ctx, w, "SUCCESS", http.StatusOK)
}

func (t *Team) Templates(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Team.Templates")
	defer span.End()

	templates := team.CustomTemplates()

	return web.Respond(ctx, w, templates, http.StatusOK)
}

func (t *Team) Modules(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Team.Modules")
	defer span.End()

	modules := team.CustomModules()

	return web.Respond(ctx, w, modules, http.StatusOK)
}

// Create inserts a new team into the system.
func (t *Team) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Team.Create")
	defer span.End()

	currentUserID, err := user.RetrieveWSCurrentUserID(ctx)
	if err != nil {
		return errors.Wrapf(err, "auth claims missing from context")
	}

	var nt team.NewTeam
	if err := web.Decode(r, &nt); err != nil {
		return errors.Wrap(err, "")
	}

	//set account_id from the request path
	nt.AccountID = params["account_id"]
	nTeam, err := team.Create(ctx, t.db, nt, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Team: %+v", &nTeam)
	}

	// create custom entities if exist.
	t.createCustomEntities(ctx, nTeam.AccountID, nTeam.ID, currentUserID, nt.Modules)

	return web.Respond(ctx, w, nTeam, http.StatusCreated)
}

func (t *Team) createCustomEntities(ctx context.Context, accountID, teamID, currentUserID string, modules []string) {

	b := base.NewBase(accountID, teamID, currentUserID, t.db, t.rPool, t.authenticator.FireBaseAdminSDK)
	b.LoadFixedEntities(ctx)

	for _, v := range modules {
		switch v {
		case "tasks":
			_, ownerSearchKey, _ := bootstrap.CurrentOwner(ctx, b.DB, b.AccountID, b.TeamID)

			_, err := b.EntityAdd(ctx, uuid.New().String(), schema.SeedTasksEntityName, "Tasks", entity.CategoryTask, entity.StateTeamLevel, forms.TaskFields(b.ContactEntity.ID, b.CompanyEntity.ID, b.OwnerEntity.ID, b.NodeEntity.ID, b.StatusEntity.ID, ownerSearchKey))
			if err != nil {
				log.Println("***> unexpected error occurred. when creating custom entity:tasks:", err)
			}
		case "deals":
			_, err := b.EntityAdd(ctx, uuid.New().String(), schema.SeedDealsEntityName, "Deals", entity.CategoryData, entity.StateTeamLevel, crm.DealFields(b.ContactEntity.ID, b.CompanyEntity.ID, b.FlowEntity.ID, b.NodeEntity.ID))
			if err != nil {
				log.Println("***> unexpected error occurred. when creating custom entity:deals:", err)
			}
		case "meetings":
			// add entity - meetings
			_, err := b.EntityAdd(ctx, uuid.New().String(), schema.SeedMeetingsEntityName, "Meetings", entity.CategoryMeeting, entity.StateTeamLevel, forms.MeetingFields(b.ContactEntity.ID, b.CompanyEntity.ID))
			if err != nil {
				log.Println("***> unexpected error occurred. when creating custom entity:meetings:", err)
			}
		case "notes":
			_, err := b.EntityAdd(ctx, uuid.New().String(), schema.SeedNotesEntityName, "Notes", entity.CategoryNotes, entity.StateTeamLevel, forms.NoteFields(b.ContactEntity.ID, b.CompanyEntity.ID))
			if err != nil {
				log.Println("***> unexpected error occurred. when creating custom entity:notes:", err)
			}
		case "items":
		case "leads":
		case "employees":
		case "tickets":
			// add entity - tickets
			_, err := b.EntityAdd(ctx, uuid.New().String(), schema.SeedTicketsEntityName, "Tickets", entity.CategoryData, entity.StateTeamLevel, forms.TicketFields(b.ContactEntity.ID, b.CompanyEntity.ID, b.StatusEntity.ID))
			if err != nil {
				log.Println("***> unexpected error occurred. when creating custom entity:tickets:", err)
			}
		}
	}

}

func (t *Team) createCustomTemplates(ctx context.Context, accountID, userID, templateKey string) error {
	switch templateKey {
	case "crm":
		err := bootstrap.BootCRM(accountID, userID, t.db, t.rPool, t.authenticator.FireBaseAdminSDK)
		if err != nil {
			log.Println("***> unexpected error occurred. when creating custom templates:crm:", err)
			return err
		}
	case "support":
		err := bootstrap.BootCRM(accountID, userID, t.db, t.rPool, t.authenticator.FireBaseAdminSDK)
		if err != nil {
			log.Println("***> unexpected error occurred. when creating custom templates:support:", err)
			return err
		}
	case "onboarding-emp":
		err := bootstrap.BootCRM(accountID, userID, t.db, t.rPool, t.authenticator.FireBaseAdminSDK)
		if err != nil {
			log.Println("***> unexpected error occurred. when creating custom templates:onboarding-emp:", err)
			return err
		}
	case "onboarding-cust":
		err := bootstrap.BootCRM(accountID, userID, t.db, t.rPool, t.authenticator.FireBaseAdminSDK)
		if err != nil {
			log.Println("***> unexpected error occurred. when creating custom templates:onboarding-cust:", err)
			return err
		}
	}
	return nil
}
