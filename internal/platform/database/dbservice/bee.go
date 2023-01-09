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
	items, err := item.Result(ctx, accountID, entityID, page, where, bee.pdb)
	if err != nil {
		return nil, nil, err
	}
	counts, err := item.Counts(ctx, accountID, entityID, where, bee.pdb)
	if err != nil {
		return items, nil, err
	}
	return items, counts, nil
}

func (bee BeeService) Count(ctx context.Context, accountID, entityID, groupByKey, groupById, groupLogic string, conditions []graphdb.Field) ([]Counters, error) {
	wh := WhBuilder(conditions)
	where := strings.Join(wh, " AND ")
	if len(wh) > 0 {
		where = fmt.Sprintf("AND %s", where)
	}
	var sel, grp string
	switch groupLogic {
	case "g_b_id": // almost all other charts from dashboard
		grpID := grpByKey(conditions)
		if grpID == "" {
			sel = "item_id, count(*)"
			grp = "group by item_id"
		} else {
			sel = fmt.Sprintf("fieldsb->>'%s' as group_id, count(*)", grpID)
			grp = "group by group_id"
		}
	case "g_b_f": // goals,activities per name
		sel = fmt.Sprintf("fieldsb->>'%s' as group_id, count(*)", groupByKey)
		grp = fmt.Sprintf("group by fieldsb->>'%s'", groupByKey)
	case "g_b_f_r": //task count per project in board view
		sel = fmt.Sprintf("genie_id as item_id, fieldsb->>'%s' as group_id, count(*)", groupByKey)
		grp = "group by item_id, group_id"
	case "g_b_f_r2": // old usage... overview
		sel = fmt.Sprintf("fieldsb->>'%s' as group_id, count(*)", groupByKey)
		grp = fmt.Sprintf("group by fieldsb->>'%s'", groupByKey)
	case "g_b_p": // project delayed
		sel = "item_id, count(*)"
		grp = "group by item_id"
	default:
		sel = "count(*)"
		grp = "group by item_id"
	}

	counts, err := item.CountMap(ctx, accountID, entityID, sel, where, grp, bee.pdb)
	if err != nil {
		return nil, err
	}

	//multiple grouping not working... making it working here
	if groupLogic == "g_b_f_r" {
		newCounter := make(map[string]int, 0)
		for _, c := range counts {
			genies := strings.Split(*c.ID, "#")
			itemID := genies[1]
			theKey := fmt.Sprintf("%s#%s", itemID, *c.GroupID)
			if v, ok := newCounter[theKey]; ok {
				newCounter[theKey] = v + 1
			} else {
				newCounter[theKey] = 1
			}
		}
		counts = make([]item.Counter, 0)
		//revert to counts
		for k, v := range newCounter {
			genies := strings.Split(k, "#")
			c := item.Counter{
				ID:      &genies[0],
				GroupID: &genies[1],
				Count:   v,
			}
			counts = append(counts, c)
		}
	}

	counters := make([]Counters, 0)
	for _, c := range counts {
		count := c.Count
		itemID := "total_count"
		if c.ID != nil {
			itemID = *c.ID
		}
		var groupID string
		if c.GroupID != nil {
			groupID = strings.ReplaceAll(*c.GroupID, "[", "")
			groupID = strings.ReplaceAll(groupID, "]", "")
			groupID = strings.ReplaceAll(groupID, "\"", "")
			slice := strings.Split(groupID, ",")
			if len(slice) > 0 {
				groupID = slice[0]
			}
			if c.ID == nil {
				itemID = groupID
			}
		}
		counters = append(counters, Counters{ID: itemID, GroupID: groupID, Count: count})
	}

	return counters, nil
}

func (bee BeeService) Sum(ctx context.Context, accountID, entityID, groupById string, conditions []graphdb.Field) ([]Counters, error) {
	return nil, nil
}

func (bee BeeService) Search1(ctx context.Context, accountID, entityID string, conditionFields []graphdb.Field) []interface{} {
	log.Printf("Search1 conditionFields %+v", conditionFields)
	return nil
}

func (bee BeeService) Search2(ctx context.Context, accountID, entityID string, conditions []graphdb.Field) ([]item.Item, error) {
	log.Printf("Search2 conditionFields %+v ", conditions)
	wh := WhBuilder(conditions)
	where := strings.Join(wh, " AND ")
	if len(wh) > 0 {
		where = fmt.Sprintf("AND %s", where)
	}
	items, err := item.Result(ctx, accountID, entityID, 0, where, bee.pdb)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (bee BeeService) Search3(ctx context.Context, accountID, entityID string, conditionFields []graphdb.Field) []interface{} {
	log.Printf("Search3 conditionFields %+v ", conditionFields)
	return nil
}

func WhBuilder(conditionFields []graphdb.Field) []string {
	wh := make([]string, 0, len(conditionFields))
	for _, f := range conditionFields {

		if f.Key == "" { // must be the source
			var value interface{}
			if f.Field != nil && f.Field.Value != nil {
				value = f.Field.Value
			} else if f.Value != nil {
				value = f.Value
			}

			switch v := value.(type) {
			case string:
				wh = append(wh, fmt.Sprintf(`genie_id = '%s#%s'`, f.RefID, v))
			case []string:
				strArr := make([]string, 0)
				for _, ev := range v {
					strArr = append(strArr, fmt.Sprintf("'%s#%s'", f.RefID, ev))
				}
				if len(strArr) > 0 {
					ids := strings.Join(strArr[:], ",")
					wh = append(wh, fmt.Sprintf(`genie_id in (%s)`, ids))
				}
			case []interface{}:
				strArr := make([]string, 0)
				for _, ev := range v {
					strArr = append(strArr, fmt.Sprintf("'%s#%s'", f.RefID, ev))
				}

				if len(strArr) > 0 {
					ids := strings.Join(strArr[:], ",")
					wh = append(wh, fmt.Sprintf(`genie_id in (%s)`, ids))
				}
			}
			continue
		}

		switch f.DataType {
		case graphdb.TypeList:
			cond := addStringifiedQuotes(f.Key, f.Field.Expression, f.Field.Value)
			if cond != "" {
				wh = append(wh, cond)
			}
		case graphdb.TypeWist: //come here when childitems
			ids := util.ConvertIntfToCommaSepString(f.Value)
			if ids != "" {
				wh = append(wh, fmt.Sprintf(`item_id in (%s)`, ids))
			} else {
				wh = append(wh, fmt.Sprintf(`item_id in (%s)`, fmt.Sprintf("'%s'", NoID))) //important. if you ignore, all child items will appear even if it is not actually associated with the item
			}
		case graphdb.TypeReference:
			cond := addStringifiedQuotes(f.Key, f.Field.Expression, f.Field.Value)
			if cond != "" {
				wh = append(wh, cond)
			}
		case graphdb.TypeString:
			if f.Expression == "STARTS WITH" {
				str := `LOWER(fieldsb->>'` + f.Key + `') LIKE LOWER('` + f.Value.(string) + `%')`
				wh = append(wh, str)
			} else {
				wh = append(wh, fmt.Sprintf("fieldsb->>'%s' %s '%s'", f.Key, f.Expression, f.Value))
			}

		case graphdb.TypeNumber:
			if f.IsDate { // the make condition field converts datetime to number for graph convinence
				wh = append(wh, fmt.Sprintf("(fieldsb->>'%s')::timestamp %s '%v'", f.Key, f.Expression, util.ConvertMilliToTimeFromIntf(f.Value)))
			} else {
				wh = append(wh, fmt.Sprintf("(fieldsb->>'%s')::int %s %v", f.Key, f.Expression, f.Value))
			}
		case graphdb.TypeDateTime: //datetime in graph DB always expects a range
			wh = append(wh, fmt.Sprintf("(fieldsb->>'%s')::timestamp %s '%v'", f.Key, f.Expression, util.ConvertMilliToTimeFromIntf(f.Value)))
		case graphdb.TypeDateTimeMillis: //datetime in graph DB always expects a range
			wh = append(wh, fmt.Sprintf("(fieldsb->>'%s')::timestamp %s '%v'", f.Key, f.Expression, util.ConvertMilliToTimeFromIntf(f.Value)))
		case graphdb.TypeDateRange:
			wh = append(wh, fmt.Sprintf("(fieldsb->>'%s')::timestamp %s '%v' AND (fieldsb->>'%s')::timestamp %s '%v'", f.Key, ">=", util.ConvertMilliToTimeFromIntf(f.Min), f.Key, "<=", util.ConvertMilliToTimeFromIntf(f.Max)))
		default:
			wh = append(wh, fmt.Sprintf("fieldsb->>'%s' %s %v", f.Key, f.Expression, f.Value))
		}
	}
	return wh
}

func grpByKey(conditionFields []graphdb.Field) string {
	var grpId string
	for _, c := range conditionFields {
		if c.Key != "" { //TODO: Bad logic.... infering group by column
			grpId = c.Key
		}
	}
	return grpId
}

func addStringifiedQuotes(key, exp string, inf interface{}) string {
	switch v := inf.(type) {
	default:
		log.Printf("inf ---- %+v", inf)
		return ""
	case string:
		q := fmt.Sprintf(`fieldsb @> '{"%s": [%s]}'`, key, fmt.Sprintf("\"%s\"", v))
		if exp == "NOT IN" {
			q = fmt.Sprintf(`NOT %s`, q)
		}
		return q
	case []string:
		s := make([]string, len(v))
		for i, v := range v {
			s[i] = fmt.Sprintf(`fieldsb @> '{"%s": [%s]}'`, key, fmt.Sprintf("\"%s\"", v))
			if exp == "NOT IN" {
				s[i] = fmt.Sprintf(`NOT %s`, s[i])
			}
		}
		return fmt.Sprintf("(%s)", strings.Join(s[:], " OR "))
	case []interface{}:
		s := make([]string, len(v))
		for i, v := range v {
			s[i] = fmt.Sprintf(`fieldsb @> '{"%s": [%s]}'`, key, fmt.Sprintf("\"%s\"", v))
			if exp == "NOT IN" {
				s[i] = fmt.Sprintf(`NOT %s`, s[i])
			}
		}
		return fmt.Sprintf("(%s)", strings.Join(s[:], " OR "))
	}
}
