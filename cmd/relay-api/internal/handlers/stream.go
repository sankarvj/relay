package handlers

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"go.opencensus.io/trace"
)

//remove not needed.....
type Stream struct {
	db *sqlx.DB
}

func (st *Stream) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Stream.List")
	defer span.End()

	accountID, _, itemID := takeAEI(ctx, params, st.db)

	e, err := entity.RetrieveFixedEntity(ctx, st.db, accountID, params["team_id"], entity.FixedEntityStream)
	if err != nil {
		return err
	}

	items, err := item.GenieEntityItems(ctx, []string{e.ID}, itemID, st.db)
	if err != nil {
		return err
	}

	fields := e.FieldsIgnoreError()
	viewModelItems := itemResponse(items)

	response := struct {
		Items  []ViewModelItem        `json:"items"`
		Fields []entity.Field         `json:"fields"`
		Entity entity.ViewModelEntity `json:"entity"`
	}{
		Items:  viewModelItems,
		Fields: fields,
		Entity: createViewModelEntity(e),
	}

	return web.Respond(ctx, w, response, http.StatusOK)
}
