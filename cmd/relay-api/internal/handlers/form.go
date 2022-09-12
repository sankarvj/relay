package handlers

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"go.opencensus.io/trace"
)

// Form represents the Form API method handler set.
type Form struct {
	db *sqlx.DB
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

	reference.UpdateReferenceFields(ctx, accountID, entityID, fields, []item.Item{it}, map[string]interface{}{}, f.db, job.NewJabEngine())

	itemMeta := it.Meta()
	itemFields := it.Fields()
	formFields := make([]entity.Field, 0)
	for _, f := range fields {
		if _, ok := itemFields[f.Key]; ok {
			formFields = append(formFields, f)
		}
	}

	formRender := struct {
		Title               string         `json:"title"`
		Description         string         `json:"description"`
		FormBackgroundColor string         `json:"form_background_color"`
		BackgroundColor     string         `json:"background_color"`
		SubmitButtonColor   string         `json:"submit_button_color"`
		SubmitButtonText    string         `json:"submit_button_text"`
		Fields              []entity.Field `json:"fields"`
	}{
		itemMeta["title"].(string),
		itemMeta["description"].(string),
		itemMeta["form_background_color"].(string),
		itemMeta["background_color"].(string),
		itemMeta["submit_button_color"].(string),
		itemMeta["submit_button_text"].(string),
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
