package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/visitor"
	"go.opencensus.io/trace"
)

// Visitor represents the Visitor API method handler set.
type Visitor struct {
	db            *sqlx.DB
	sdb           *database.SecDB
	authenticator *auth.Authenticator
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing items
func (v *Visitor) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Visitor.List")
	defer span.End()

	accountID := params["account_id"]
	page := util.ConvertStrToInt(r.URL.Query().Get("page"))

	visitors, err := visitor.List(ctx, accountID, page, v.db)
	if err != nil {
		return err
	}

	entityIds := make([]string, 0)
	for _, v := range visitors {
		entityIds = append(entityIds, v.EntityID)
	}

	associatedEntities, err := entity.BulkRetrieve(ctx, entityIds, v.db)
	if err != nil {
		return err
	}

	entityMap := make(map[string]string, 0)
	for _, e := range associatedEntities {
		entityMap[e.ID] = e.DisplayName
	}

	viewModelVisitors := make([]ViewModelVisitor, 0)
	for i := 0; i < len(visitors); i++ {
		entityName := entityMap[visitors[i].EntityID]
		viewModelVisitors = append(viewModelVisitors, createViewModelVisitor(visitors[i], entityName, ""))
	}

	return web.Respond(ctx, w, viewModelVisitors, http.StatusOK)
}

func (v *Visitor) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Visitor.Retrieve")
	defer span.End()

	accountID := params["account_id"]
	visitorID := params["visitor_id"]

	visitor, err := visitor.Retrieve(ctx, accountID, visitorID, v.db)
	if err != nil {
		return err
	}

	e, err := entity.Retrieve(ctx, accountID, visitor.EntityID, v.db, v.sdb)
	if err != nil {
		return err
	}

	i, err := item.Retrieve(ctx, accountID, visitor.EntityID, visitor.ItemID, v.db)
	if err != nil {
		return err
	}

	valueAddedFields := e.ValueAdd(i.Fields())
	namedActualFields := entity.MetaMap(valueAddedFields)
	title := namedActualFields["title"].Value.(string)

	return web.Respond(ctx, w, createViewModelVisitor(visitor, e.DisplayName, title), http.StatusOK)
}

func (v *Visitor) RetrieveItem(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Visitor.List")
	defer span.End()

	mlToken := r.URL.Query().Get("ml_token")

	userInfo, err := auth.AuthenticateToken(mlToken, v.sdb)
	if err != nil {
		return errors.Wrap(err, "verifying mlToken")
	}

	vis, err := visitor.Retrieve(ctx, userInfo.AccountID, userInfo.MemberID, v.db)
	if err != nil {
		return err
	}

	if vis.Token != mlToken {
		return errors.Wrap(err, "token mismatch detected")
	}

	params["account_id"] = vis.AccountID
	params["team_id"] = vis.TeamID
	params["entity_id"] = vis.EntityID
	params["item_id"] = vis.ItemID

	i := Item{
		db:            v.db,
		sdb:           v.sdb,
		authenticator: v.authenticator,
	}

	return web.Respond(ctx, w, i.List(ctx, w, r, params), http.StatusOK)
}

//Update updates the item
func (v *Visitor) ToggleActive(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Visitor.ToggleActive")
	defer span.End()

	accountID := params["account_id"]
	visitorID := params["visitor_id"]

	vis, err := visitor.Retrieve(ctx, accountID, visitorID, v.db)
	if err != nil {
		return err
	}

	err = visitor.UpdateActive(ctx, v.db, accountID, visitorID, !vis.Active)
	if err != nil {
		return errors.Wrapf(err, "Cannot change active state for this visitor")
	}

	return web.Respond(ctx, w, "SUCCESS", http.StatusOK)
}

func (v *Visitor) Resend(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Visitor.Resend")
	defer span.End()

	accountID := params["account_id"]
	visitorID := params["visitor_id"]

	var vmv ViewModelVisitor
	if err := web.Decode(r, &vmv); err != nil {
		return errors.Wrap(err, "")
	}

	err := visitor.UpdateExpiry(ctx, v.db, accountID, visitorID)
	if err != nil {
		return errors.Wrapf(err, "Cannot change active state for this visitor")
	}

	go job.NewJob(v.db, v.sdb, v.authenticator.FireBaseAdminSDK).AddVisitor(accountID, visitorID, vmv.Body, v.db, v.sdb)

	return web.Respond(ctx, w, "SUCCESS", http.StatusOK)
}

func (v *Visitor) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Visitor.Delete")
	defer span.End()

	accountID := params["account_id"]
	visitorID := params["visitor_id"]

	err := visitor.Delete(ctx, v.db, accountID, visitorID)
	if err != nil {
		return errors.Wrapf(err, "Cannot delete this visitor")
	}

	return web.Respond(ctx, w, "SUCCESS", http.StatusOK)
}

// Create inserts a new team into the system.
func (v *Visitor) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Visitor.Create")
	defer span.End()
	accountID := params["account_id"]

	var vmv ViewModelVisitor
	if err := web.Decode(r, &vmv); err != nil {
		return errors.Wrap(err, "")
	}

	token, _ := auth.GenerateRandomToken(32)
	nv := visitor.NewVisitor{
		AccountID: accountID,
		TeamID:    vmv.TeamID,
		EntityID:  vmv.EntityID,
		ItemID:    vmv.ItemID,
		Name:      vmv.Name,
		Email:     vmv.Email,
		Token:     token,
	}
	vis, err := visitor.Create(ctx, v.db, nv, time.Now())
	if err != nil {
		return err
	}

	go job.NewJob(v.db, v.sdb, v.authenticator.FireBaseAdminSDK).AddVisitor(vis.AccountID, vis.VistitorID, vmv.Body, v.db, v.sdb)

	return web.Respond(ctx, w, createViewModelVisitor(vis, vmv.EntityName, vmv.ItemName), http.StatusCreated)
}

func (v *Visitor) ListRelations(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Visitor.List")
	defer span.End()
	response := ""
	return web.Respond(ctx, w, response, http.StatusOK)
}

func (v *Visitor) ChildItems(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Visitor.ChildItems")
	defer span.End()
	response := ""
	return web.Respond(ctx, w, response, http.StatusOK)
}
