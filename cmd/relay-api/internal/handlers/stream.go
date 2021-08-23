package handlers

import (
	"context"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"go.opencensus.io/trace"
)

type Stream struct {
	db    *sqlx.DB
	rPool *redis.Pool
}

func (st *Stream) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Stream.List")
	defer span.End()

	accountID, _, itemID := takeAEI(ctx, params, st.db)

	e, err := entity.RetrieveFixedEntity(ctx, st.db, accountID, entity.FixedEntityStream)
	if err != nil {
		return err
	}

	items, err := item.GenieEntityItems(ctx, e.ID, itemID, st.db)
	if err != nil {
		return err
	}

	fields, viewModelItems := itemResponse(e, items)

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
