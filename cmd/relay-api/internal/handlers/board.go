package handlers

import (
	"context"
	"strings"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/reference"
)

func groupBy(ctx context.Context, fieldKey string, enty entity.Entity, db *sqlx.DB) ([]reference.Choicer, error) {
	var choicers []reference.Choicer
	var groupByField *entity.Field
	fields := enty.FieldsIgnoreError()

	for _, f := range fields {
		if f.Key == fieldKey {
			groupByField = &f
			break
		}
	}

	if groupByField != nil && groupByField.IsReference() {
		e, err := entity.Retrieve(ctx, enty.AccountID, groupByField.RefID, db)
		if err != nil {
			return choicers, err
		}
		items, err := item.EntityItems(ctx, groupByField.RefID, db)
		if err != nil {
			return choicers, err
		}
		choicers = reference.ItemChoices(groupByField, items, e.WhoFields())
	}
	choicers = append(choicers, reference.Choicer{ID: "", Name: "Others", Value: itIds(choicers)})
	return choicers, nil
}

func itIds(choicers []reference.Choicer) string {
	ids := make([]string, len(choicers))
	for index, it := range choicers {
		ids[index] = it.ID
	}
	return strings.Join(ids[:], ",")
}
