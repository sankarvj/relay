package event

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/timeseries"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

type TSData struct {
	UpdateExisting bool
	NewData        *timeseries.Timeseries
	OldData        *timeseries.Timeseries
}

func Process(ctx context.Context, accountID string, ne NewEvent, log *log.Logger, db *sqlx.DB) (item.Item, error) {
	e, err := entity.RetrieveByName(ctx, accountID, ne.Block, db)
	if err != nil {
		log.Println("processEvent : errored : save event : retrive entity")
		return item.Item{}, err
	}

	var name *string
	if ne.Properties["name"] != nil {
		namePtr := ne.Properties["name"].(string)
		name = &namePtr
	}

	fields := e.KeyMap(ne.Properties)
	//adding identifier as one of the field inside the item
	fields["identifier"] = ne.Identifier

	userID := user.UUID_SYSTEM_USER
	ni := item.NewItem{
		ID:        uuid.New().String(),
		Name:      name,
		AccountID: accountID,
		EntityID:  e.ID,
		UserID:    &userID,
		Fields:    fields,
		Source:    nil,
	}

	it, err := item.Create(ctx, db, ni, time.Now())
	return it, err
}

// func Process(ctx context.Context, accountID string, ne NewEvent, log *log.Logger, db *sqlx.DB) (TSData, error) {
// 	tsData := TSData{}
// 	now := time.Now()
// 	//actual entity : page_view, events, errors, sign_ups, subscriptions
// 	e, err := entity.RetrieveByName(ctx, accountID, ne.Block, db)
// 	if err != nil {
// 		log.Println("processEvent : errored : save event : retrive entity")
// 		return tsData, err
// 	}

// 	var iMap map[string]interface{}
// 	inrec, _ := json.Marshal(ne)
// 	json.Unmarshal(inrec, &iMap)

// 	data, err := timeseries.RetriveLatest(ctx, accountID, e.ID, tsIdentifier(ne.Tags), db)
// 	if err != nil {
// 		log.Println("processEvent : errored : save event : retrive items")
// 		return tsData, err
// 	}

// 	oldTS := timeseries.Timeseries{}
// 	if len(data) > 0 {
// 		oldTS = data[0]
// 	}

// 	tsFields := oldTS.Fields()
// 	tsNamedFields := make(map[string]interface{}, 0)
// 	namedFieldsMap := e.NameMapWrapper()
// 	for name, v := range body {
// 		if f, ok := namedFieldsMap[name]; ok {
// 			if f.Who == entity.WhoCounter {
// 				rollUp(f)
// 				tsFields[f.Key] = v
// 				tsNamedFields[f.Name] = v
// 			}
// 		}
// 	}

// 	if tsData.UpdateExisting && oldTS.ID != "" {
// 		tsData.OldData = &oldTS
// 		newTS, err := updateTimeseries(ctx, db, oldTS, tsFields, tsNamedFields, now)
// 		if err != nil {
// 			log.Println("processEvent : errored : save event")
// 			return tsData, err
// 		}
// 		tsData.NewData = newTS
// 		log.Println("processEvent : completed : save event")

// 		return tsData, nil
// 	} else {
// 		tsData.NewData, err = createTimeseries(ctx, db, accountID, e.ID, tsFields, tsNamedFields, now)
// 		if err != nil {
// 			log.Println("processEvent : errored : save event")
// 			return tsData, err
// 		}
// 		log.Println("processEvent : completed : save event")
// 		return tsData, nil
// 	}

// }

// func rollUp(f entity.Field, tsData *TSData) {
// 	switch f.RollUp() {
// 	case entity.MetaRollUpAlways:
// 		tsData.UpdateExisting = true
// 		v = f.CalcFunc().Calc(tsFields[f.Key], v)
// 	case entity.MetaRollUpNever:
// 		tsData.UpdateExisting = false
// 	case entity.MetaRollUpHourly:
// 		if oldTS.ID != "" && hourEqual(oldTS.EndTime, now) {
// 			tsData.UpdateExisting = true
// 			v = f.CalcFunc().Calc(tsFields[f.Key], v)
// 		}
// 	case entity.MetaRollUpDaily:
// 		if oldTS.ID != "" && dateEqual(oldTS.EndTime, now) {
// 			tsData.UpdateExisting = true
// 			v = f.CalcFunc().Calc(tsFields[f.Key], v)
// 		}
// 	case entity.MetaRollUpMinute:
// 		if oldTS.ID != "" && minuteEqual(oldTS.EndTime, now) {
// 			tsData.UpdateExisting = true
// 			v = f.CalcFunc().Calc(tsFields[f.Key], v)
// 		}
// 	case entity.MetaRollUpChangeOver:
// 		if oldTS.ID != "" && oldTS.Count != v {
// 			tsData.UpdateExisting = true
// 		}
// 	}
// }

// func createTimeseries(ctx context.Context, db *sqlx.DB, accountID, entityID string, keyedFields, namedFields map[string]interface{}, now time.Time) (*timeseries.Timeseries, error) {
// 	nt := timeseries.NewTimeseries{
// 		ID:          uuid.New().String(),
// 		AccountID:   accountID,
// 		EntityID:    entityID,
// 		Type:        timeseries.TypeUnknown,
// 		Event:       tsEvent(namedFields),
// 		Description: tsDesc(namedFields),
// 		Count:       tsCount(namedFields),
// 		Tags:        tsTags(namedFields),
// 		Identifier:  tsIdentifier(namedFields),
// 		Fields:      keyedFields,
// 	}
// 	t, err := timeseries.Create(ctx, db, nt, now)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &t, nil
// }

// func updateTimeseries(ctx context.Context, db *sqlx.DB, oldTS timeseries.Timeseries, keyedFields, namedFields map[string]interface{}, now time.Time) (*timeseries.Timeseries, error) {
// 	fieldsBytes, err := json.Marshal(keyedFields)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "encode fields to bytes")
// 	}

// 	oldTS.Event = tsEvent(namedFields)
// 	oldTS.Description = tsDesc(namedFields)
// 	oldTS.Count = tsCount(namedFields)
// 	oldTS.Tags = tsTags(namedFields)
// 	oldTS.Identifier = tsIdentifier(namedFields)
// 	oldTS.Fieldsb = string(fieldsBytes)

// 	t, err := timeseries.Update(ctx, db, oldTS, now)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return t, nil
// }

func tsEvent(namedFields map[string]interface{}) string {
	if val, ok := namedFields["event"]; ok {
		return val.(string)
	}
	return ""
}

func tsDesc(namedFields map[string]interface{}) string {
	if val, ok := namedFields["description"]; ok {
		return val.(string)
	}
	return ""
}

func tsCount(namedFields map[string]interface{}) int {
	if val, ok := namedFields["count"]; ok {
		_, ok = val.(int)
		if !ok {
			_, ok = val.(float64)
			if !ok {
				return util.ConvertStrToInt(val.(string))
			} else {
				return int(val.(float64))
			}
		} else {
			return val.(int)
		}
	}
	return 0
}

func tsTags(namedFields map[string]interface{}) []string {
	if val, ok := namedFields["tags"]; ok {
		return util.ConvertSliceTypeRev(val.([]interface{}))
	}
	return []string{}
}

func dateEqual(time1, time2 time.Time) bool {
	y1, m1, d1 := time1.Date()
	y2, m2, d2 := time2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func hourEqual(time1, time2 time.Time) bool {
	h1 := time1.Hour()
	h2 := time2.Hour()
	return h1 == h2 && dateEqual(time1, time2)
}

func minuteEqual(time1, time2 time.Time) bool {
	m1 := time1.Minute()
	m2 := time2.Minute()
	return m1 == m2 && hourEqual(time1, time2)
}
