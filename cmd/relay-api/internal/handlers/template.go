package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
)

func (i *Item) CreateTemplate(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Template.Create")
	defer span.End()

	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return err
	}

	var ni item.NewItem
	if err := web.Decode(r, &ni); err != nil {
		return errors.Wrap(err, "")
	}

	//current entity
	ce, err := entity.Retrieve(ctx, params["account_id"], params["entity_id"], i.db)
	if err != nil {
		return err
	}

	ni.AccountID = params["account_id"]
	ni.EntityID = params["entity_id"]
	ni.UserID = &currentUserID
	ni.ID = uuid.New().String()
	ni.State = item.StateBluePrint
	valueAddedFields := ce.ValueAdd(ni.Fields)

	for _, f := range valueAddedFields {
		if f.IsTitleLayout() {
			s := f.Value.(string)
			ni.Name = &s
		}

		if f.IsDateTime() {
			ni.Fields[f.Key] = fmt.Sprintf("<<%v>>", f.Value)
		}
	}

	it, err := item.Create(ctx, i.db, ni, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Item: %+v", &i)
	}

	return web.Respond(ctx, w, createViewModelItem(it), http.StatusCreated)
}