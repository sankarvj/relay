package dbservice

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

const (
	NoID = "00000000-0000-0000-0000-000000000000"
)

type BeeService struct {
	pdb *sqlx.DB
	sdb *database.SecDB
}

func (bee BeeService) Result(ctx context.Context, accountID, entityID, sortby, direction string, page int, docount, useReturn bool, conditions []graphdb.Field) ([]item.Item, map[string]int, error) {
	wh := WhBuilder(conditions)
	where := strings.Join(wh, " AND ")
	if len(wh) > 0 {
		where = fmt.Sprintf("AND %s", where)
	}
	items, err := item.Result(ctx, accountID, entityID, 0, where, bee.pdb)
	if err != nil {
		return nil, nil, err
	}
	counts, err := item.Counts(ctx, accountID, entityID, where, bee.pdb)
	if err != nil {
		return items, nil, err
	}
	return items, counts, nil
}

func (bee BeeService) Count(ctx context.Context, accountID, entityID, groupById, groupLogic string, conditions []graphdb.Field) ([]Counters, error) {
	return nil, nil
}

func (bee BeeService) Sum(ctx context.Context, accountID, entityID, groupById string, conditions []graphdb.Field) ([]Counters, error) {
	return nil, nil
}

func (bee BeeService) Search1(ctx context.Context, accountID, entityID string, conditionFields []graphdb.Field) []interface{} {
	return nil
}

func (bee BeeService) Search2(ctx context.Context, accountID, entityID string, conditionFields []graphdb.Field) ([]item.Item, error) {
	return nil, nil
}

func (bee BeeService) Search3(ctx context.Context, accountID, entityID string, conditionFields []graphdb.Field) []interface{} {
	return nil
}

func WhBuilder(conditionFields []graphdb.Field) []string {
	wh := make([]string, 0, len(conditionFields))
	for _, f := range conditionFields {

		log.Printf("f::::f:::::f::::::f -------> %+v", f)

		if f.Key == "" { // must be the source
			wh = append(wh, fmt.Sprintf(`genie_id = '%s#%s'`, f.RefID, f.Field.Value))
			continue
		}

		switch f.DataType {
		case graphdb.TypeList:
			wh = append(wh, fmt.Sprintf(`fieldsb @> '{"%s": %s}'`, f.Key, util.AddStringifiedQuotes(f.Field.Value)))
		case graphdb.TypeWist: //come here when childitems
			ids := util.ConvertIntfToCommaSepString(f.Value)
			if ids != "" {
				wh = append(wh, fmt.Sprintf(`item_id in (%s)`, ids))
			} else {
				wh = append(wh, fmt.Sprintf(`item_id in (%s)`, fmt.Sprintf("'%s'", NoID))) //important. if you ignore, all child items will appear even if it is not actually associated with the item
			}
		case graphdb.TypeReference:
			log.Printf("fFieldTypeReference-------> %+v", f.Field)
			wh = append(wh, fmt.Sprintf(`fieldsb @> '{"%s": %s}'`, f.Key, util.AddStringifiedQuotes(f.Field.Value)))
		case graphdb.TypeString:
			wh = append(wh, fmt.Sprintf("fieldsb->>'%s' %s '%s'", f.Key, f.Expression, f.Value))
		case graphdb.TypeNumber:
			if f.IsDate { // the make condition field converts datetime to number for graph convinence
				wh = append(wh, fmt.Sprintf("(fieldsb->>'%s')::timestamp %s '%v'", f.Key, f.Expression, util.ConvertMilliToTimeFromIntf(f.Value)))
			} else {
				wh = append(wh, fmt.Sprintf("(fieldsb->>'%s')::int %s %v", f.Key, f.Expression, f.Value))
			}
		case graphdb.TypeDateTime: //datetime in graph DB always expects a range
			log.Printf("fTypeDateTime-------> %+v", f)
			wh = append(wh, fmt.Sprintf("(fieldsb->>'%s')::timestamp %s '%v'", f.Key, f.Expression, util.ConvertMilliToTimeFromIntf(f.Value)))
		case graphdb.TypeDateTimeMillis: //datetime in graph DB always expects a range
			log.Printf("fTypeDateTimeMillis-------> %+v", f)
			wh = append(wh, fmt.Sprintf("(fieldsb->>'%s')::timestamp %s '%v'", f.Key, f.Expression, util.ConvertMilliToTimeFromIntf(f.Value)))
		case graphdb.TypeDateRange:
			log.Printf("fTypeDateRange-------> %+v", f)
			wh = append(wh, fmt.Sprintf("(fieldsb->>'%s')::timestamp %s '%v' AND (fieldsb->>'%s')::timestamp %s '%v'", f.Key, ">=", util.ConvertMilliToTimeFromIntf(f.Min), f.Key, "<=", util.ConvertMilliToTimeFromIntf(f.Max)))
		default:
			wh = append(wh, fmt.Sprintf("fieldsb->>'%s' %s %v", f.Key, f.Expression, f.Value))
		}
	}
	return wh
}
