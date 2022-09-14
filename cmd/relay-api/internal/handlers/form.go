package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
)

// Form represents the Form API method handler set.
type Form struct {
	db            *sqlx.DB
	sdb           *database.SecDB
	authenticator *auth.Authenticator
}

func (f *Form) Adder(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Form.Create")
	defer span.End()

	accountID, entityID, itemID := takeAEI(ctx, params, f.db)

	var newItem item.NewItem
	if err := web.Decode(r, &newItem); err != nil {
		return errors.Wrap(err, "")
	}

	e, err := entity.Retrieve(ctx, accountID, entityID, f.db)
	if err != nil {
		return err
	}
	fields := e.FieldsIgnoreError()

	it, err := item.Retrieve(ctx, entityID, itemID, f.db)
	if err != nil {
		return err
	}

	if it.State != item.StateWebForm {
		return web.NewRequestError(err, http.StatusForbidden)
	}

	if validate(fields, it.Fields(), newItem.Fields) != nil {
		web.NewRequestError(err, http.StatusBadRequest)
	}

	anonymousUser := user.UUID_ANONYMOUS_USER
	ni := item.NewItem{
		ID:        uuid.New().String(),
		Name:      it.Name,
		AccountID: accountID,
		EntityID:  entityID,
		GenieID:   &it.ID,
		UserID:    &anonymousUser,
		Fields:    newItem.Fields,
		Source:    map[string][]string{entityID: {it.ID}},
		Type:      item.TypeForm,
	}

	_, err = createAndPublish(ctx, anonymousUser, ni, f.db, f.sdb, f.authenticator.FireBaseAdminSDK)
	if err != nil {
		return errors.Wrapf(err, "Form: %+v", &ni)
	}

	return web.Respond(ctx, w, "success", http.StatusCreated)
}

func (f *Form) Render(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Form.Render")
	defer span.End()

	accountID, entityID, itemID := takeAEI(ctx, params, f.db)
	e, err := entity.Retrieve(ctx, accountID, entityID, f.db)
	if err != nil {
		return err
	}
	fields := e.FieldsIgnoreError()

	it, err := item.Retrieve(ctx, entityID, itemID, f.db)
	if err != nil {
		return err
	}

	if it.State != item.StateWebForm {
		return web.NewRequestError(err, http.StatusForbidden)
	}

	reference.UpdateReferenceFields(ctx, accountID, entityID, fields, []item.Item{}, map[string]interface{}{}, f.db, job.NewJabEngine())

	itemMeta := it.Meta()
	itemFields := it.Fields()
	formFields := make([]entity.Field, 0)
	for _, f := range fields {
		if itfm, ok := itemFields[f.Key]; ok {
			f.Meta = convertIntfMaptoStrMap(itfm.(map[string]interface{}))
			formFields = append(formFields, f)
		}
	}

	formRender := struct {
		Title               interface{}    `json:"title"`
		Description         interface{}    `json:"description"`
		FormBackgroundColor interface{}    `json:"form_background_color"`
		BackgroundColor     interface{}    `json:"background_color"`
		SubmitButtonColor   interface{}    `json:"submit_button_color"`
		SubmitButtonText    interface{}    `json:"submit_button_text"`
		Fields              []entity.Field `json:"fields"`
	}{
		itemMeta["title"],
		itemMeta["description"],
		itemMeta["form_background_color"],
		itemMeta["background_color"],
		itemMeta["submit_button_color"],
		itemMeta["submit_button_text"],
		formFields,
	}

	return web.Respond(ctx, w, formRender, http.StatusOK)
}

func (f *Form) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Form.List")
	defer span.End()

	page := util.ConvertStrToInt(r.URL.Query().Get("page"))

	entities, err := entity.TeamEntities(ctx, params["account_id"], params["team_id"], []int{}, f.db)
	if err != nil {
		return err
	}
	entityIDs := entity.FetchIDs(entities)

	forms, err := item.Forms(ctx, params["account_id"], entityIDs, page, f.db)
	if err != nil {
		return errors.Wrap(err, "retriving forms")
	}

	totalCount, _ := item.Count(ctx, params["account_id"], entityIDs, f.db)

	response := struct {
		Items      []item.Form `json:"forms"`
		TotalCount int         `json:"total_count"`
	}{
		Items:      forms,
		TotalCount: totalCount,
	}

	return web.Respond(ctx, w, response, http.StatusOK)
}

func convertIntfMaptoStrMap(data map[string]interface{}) map[string]string {
	meta := make(map[string]string)
	for k, v := range data {
		meta[k] = v.(string)
	}
	return meta
}

func validate(fields []entity.Field, itemFields map[string]interface{}, values map[string]interface{}) error {
	if len(itemFields) == 0 {
		return nil
	}

	for _, f := range fields {
		if itfm, ok := itemFields[f.Key]; ok {
			validator := convertIntfMaptoStrMap(itfm.(map[string]interface{}))
			if validator["required"] == "true" && (values[f.Key] == "" || values[f.Key] == nil) {
				return errors.New("Required field is empty")
			}
		}
	}
	return nil
}
