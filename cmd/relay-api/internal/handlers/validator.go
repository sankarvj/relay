package handlers

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/database/dbservice"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

func validateItemCreate(ctx context.Context, accountID, entityID string, values map[string]interface{}, db *sqlx.DB, sdb *database.SecDB) *ErrorResponse {
	e, err := entity.Retrieve(ctx, accountID, entityID, db, sdb)
	if err != nil {
		return unexpectedError(errors.Wrapf(err, "item create validation failed"))
	}

	//required field error
	requiredErrorsMap, err := validateRequired(ctx, e, values)
	if err != nil {
		return unexpectedError(errors.Wrapf(err, "required field validation failed"))
	}
	if len(requiredErrorsMap) > 0 {
		return requiredError(requiredErrorsMap)
	}

	//unique field error
	uniqueErrorsMap, _, err := validateUniquness(ctx, e, values, db, sdb)
	if err != nil {
		return unexpectedError(errors.Wrapf(err, "uniquness validation failed"))
	}
	if len(uniqueErrorsMap) > 0 {
		return validationError(uniqueErrorsMap)
	}

	return nil
}

func validateItemUpdate(ctx context.Context, accountID, entityID, itemID string, values map[string]interface{}, db *sqlx.DB, sdb *database.SecDB) *ErrorResponse {
	e, err := entity.Retrieve(ctx, accountID, entityID, db, sdb)
	if err != nil {
		return unexpectedError(errors.Wrapf(err, "item update validation failed"))
	}

	//required field error
	requiredErrorsMap, err := validateRequired(ctx, e, values)
	if err != nil {
		return unexpectedError(errors.Wrapf(err, "required field validation failed"))
	}
	if len(requiredErrorsMap) > 0 {
		return requiredError(requiredErrorsMap)
	}

	//unique field error
	uniqueErrorsMap, erroredIds, err := validateUniquness(ctx, e, values, db, sdb)
	if err != nil {
		return unexpectedError(errors.Wrapf(err, "uniquness validation failed"))
	}
	if len(uniqueErrorsMap) > 0 && !uniqueErrorIsInvalid(erroredIds, itemID) {
		return validationError(uniqueErrorsMap)
	}

	return nil
}

func validateRequired(ctx context.Context, e entity.Entity, values map[string]interface{}) (map[string]ErrorPayload, error) {
	fields := e.OnlyRequiredFields()
	if len(fields) == 0 {
		return nil, nil
	}

	requiredErrorsMap := make(map[string]ErrorPayload, 0)
	for _, f := range fields {
		switch values[f.Key].(type) {
		case []interface{}:
			if len(values[f.Key].([]interface{})) == 0 {
				requiredErrorsMap[f.Key] = requiredErrorPayload()
			}
		case interface{}:
			if values[f.Key] == nil || values[f.Key] == "" {
				requiredErrorsMap[f.Key] = requiredErrorPayload()
			}
		}
	}
	return requiredErrorsMap, nil
}

func validateUniquness(ctx context.Context, e entity.Entity, values map[string]interface{}, db *sqlx.DB, sdb *database.SecDB) (map[string]ErrorPayload, []string, error) {

	uniqueItems := make([]string, 0)
	//unique fields only
	fields := e.OnlyUniqueFields()
	if len(fields) == 0 {
		return nil, nil, nil
	}

	conditionFields := make([]graphdb.Field, 0)
	for _, f := range fields {

		if _, ok := values[f.Key]; !ok {
			continue
		}

		var gf graphdb.Field
		if f.IsReference() {
			gf = graphdb.Field{
				Value:    values[f.Key],
				RefID:    f.RefID,
				DataType: graphdb.DType(f.DataType),
			}
		} else {
			gf = graphdb.Field{
				Expression: "=",
				Key:        f.Key,
				DataType:   graphdb.DType(f.DataType),
				Value:      values[f.Key],
			}
		}
		conditionFields = append(conditionFields, gf)
	}

	useDB := account.UseDB(ctx, db, e.AccountID)
	items, _, err := dbservice.NewDBservice(useDB, db, sdb).Result(ctx, e.AccountID, e.ID, "", "", 0, false, false, conditionFields)
	if err != nil {
		return nil, nil, err
	}

	uniqueErrorsMap := make(map[string]ErrorPayload, 0)
	for _, i := range items {
		uniqueItems = append(uniqueItems, i.ID)
		filtertedVals := i.Fields()
		for _, f := range fields {
			switch filtertedVals[f.Key].(type) {
			case []interface{}:
				givenList := values[f.Key].([]interface{})
				dbList := filtertedVals[f.Key].([]interface{})
				uniqueErrorsMap[f.Key] = uniqueErrorPayloads(util.Similar(givenList, dbList))
			case interface{}:
				if filtertedVals[f.Key] == values[f.Key] {
					uniqueErrorsMap[f.Key] = uniqueErrorPayload(values[f.Key])
				}
			}
		}
	}
	return uniqueErrorsMap, uniqueItems, nil
}
