package mid

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"go.opencensus.io/trace"
)

func HasSeat(db *sqlx.DB) web.Middleware {

	// This is the actual middleware function to be executed.
	f := func(after web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.HasSeat")
			defer span.End()

			accountID := params["account_id"]
			acc, err := account.Retrieve(ctx, db, accountID)
			if err != nil {
				err := errors.New("acc_not_exist") // value used in the UI dont change the string message.
				return web.NewRequestError(err, http.StatusForbidden)
			}

			log.Println("Look for checking the feature access and the quantity access", acc)

			return after(ctx, w, r, params)
		}

		return h
	}

	return f
}

func plans() {
	pro := []string{"visitor", "forms", "custom_entities"}
	startup := []string{"workflow", "pipeline", "custom_entities"}
	free := []string{"pipeline"}
	log.Println(pro, startup, free)
}
