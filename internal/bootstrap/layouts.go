package bootstrap

import (
	"context"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func addLayouts(ctx context.Context, db *sqlx.DB, name, accountID, entityID string) error {
	e, err := entity.Retrieve(ctx, accountID, entityID, db)
	if err != nil {
		return err
	}
	layoutFields := make(map[string]string, 0)
	for _, f := range e.FieldsIgnoreError() {
		if f.Key == "uuid-00-name" { // confusing? because this should happen only via the UI.
			layoutFields["title"] = f.Key
		} else if f.Key == "uuid-00-owner" {
			layoutFields["owner"] = f.Key
		}
	}
	return BootstrapLayout(ctx, db, name, accountID, entityID, layoutFields)
}
