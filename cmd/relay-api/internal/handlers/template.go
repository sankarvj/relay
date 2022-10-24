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

	accountID, entityID, _ := takeAEI(ctx, params, i.db)
	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return err
	}

	var ni item.NewItem
	if err := web.Decode(r, &ni); err != nil {
		return errors.Wrap(err, "")
	}

	//current entity
	ce, err := entity.Retrieve(ctx, accountID, entityID, i.db, i.sdb)
	if err != nil {
		return err
	}

	ni.ID = uuid.New().String()
	ni.AccountID = accountID
	ni.EntityID = entityID
	ni.UserID = &currentUserID
	ni.State = item.StateBluePrint
	valueAddedFields := ce.ValueAdd(ni.Fields)

	for _, f := range valueAddedFields {
		if f.IsTitleLayout() {
			s := f.Value.(string)
			ni.Name = &s
		}

		if f.IsDateOrTime() {
			ni.Fields[f.Key] = fmt.Sprintf("<<%v>>", f.Value)
		}
	}

	it, err := item.Create(ctx, i.db, ni, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Item: %+v", &i)
	}

	return web.Respond(ctx, w, createViewModelItem(it), http.StatusCreated)
}
