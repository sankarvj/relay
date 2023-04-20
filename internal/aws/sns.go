package aws

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/database/dbservice"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/timeseries"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

// Subscription decribes the aws sns object
type Subscription struct {
	Type             string    `json:"Type"`
	MessageID        string    `json:"MessageId"`
	Token            string    `json:"Token"`
	TopicArn         string    `json:"TopicArn"`
	Subject          string    `json:"Subject"`
	Message          string    `json:"Message"`
	SubscribeURL     string    `json:"SubscribeURL"`
	Timestamp        time.Time `json:"Timestamp"`
	SignatureVersion string    `json:"SignatureVersion"`
	Signature        string    `json:"Signature"`
	SigningCertURL   string    `json:"SigningCertURL"`
	UnsubscribeURL   string    `json:"UnsubscribeURL"`
}

// Message descripbes sns message
type Message struct {
	AlarmName        string `json:"AlarmName"`
	AlarmDescription string `json:"AlarmDescription"`
	AWSAccountID     string `json:"AWSAccountId"`
	NewStateValue    string `json:"NewStateValue"`
	NewStateReason   string `json:"NewStateReason"`
	StateChangeTime  string `json:"StateChangeTime"`
	Region           string `json:"Region"`
	OldStateValue    string `json:"OldStateValue"`
	Trigger          struct {
		MetricName                       string      `json:"MetricName"`
		Namespace                        string      `json:"Namespace"`
		StatisticType                    string      `json:"StatisticType"`
		Statistic                        string      `json:"Statistic"`
		Unit                             interface{} `json:"Unit"`
		Dimensions                       []Dimension `json:"Dimensions"`
		Period                           int         `json:"Period"`
		EvaluationPeriods                int         `json:"EvaluationPeriods"`
		ComparisonOperator               string      `json:"ComparisonOperator"`
		Threshold                        float64     `json:"Threshold"`
		TreatMissingData                 string      `json:"TreatMissingData"`
		EvaluateLowSampleCountPercentile string      `json:"EvaluateLowSampleCountPercentile"`
	} `json:"Trigger"`
}

// Dimension describes sns key values
type Dimension struct {
	Value string `json:"value"`
	Name  string `json:"name"`
}

func SaveAlert(ctx context.Context, accountID string, namedFieldsMap map[string]interface{}, db *sqlx.DB, sdb *database.SecDB, fbSDKPath string) error {
	uniqueID := namedFieldsMap["unique"]
	entityID := namedFieldsMap["block"]

	if entityID == nil || entityID == "" {
		return errors.New("paramater block is missing")
	}

	e, err := entity.Retrieve(ctx, accountID, entityID.(string), db, sdb)
	if err != nil {
		return err
	}

	if e.FlowField() != nil {
		flows, err := flow.List(ctx, []string{e.ID}, flow.FlowModePipeLine, flow.FlowTypeEventCreate, db)
		if err != nil {
			return err
		}
		if len(flows) > 0 {
			namedFieldsMap[e.FlowField().Name] = []interface{}{flows[0].ID}
		}
	}

	fields := keyMap(e.NameKeyMapWrapper(), namedFieldsMap)

	itemID, err := findDuplicate(ctx, accountID, entityID, uniqueID, db, sdb)
	if err != nil {
		return err
	}

	if itemID != nil {
		return createItemDupEvent(ctx, accountID, entityID.(string), itemID, db)
	} else {
		return createItem(ctx, accountID, entityID.(string), fields, db, sdb, fbSDKPath)
	}

}

func createItemDupEvent(ctx context.Context, accountID, entityID string, itemID *string, db *sqlx.DB) error {
	log.Println("Coming to createItemDupEvent")
	now := time.Now()
	nt := timeseries.NewTimeseries{
		ID:          uuid.New().String(),
		AccountID:   accountID,
		EntityID:    entityID,
		Type:        timeseries.TypeIncident,
		Event:       "AWS incident",
		Description: "",
		Count:       1,
		Tags:        []string{},
		Identifier:  itemID,
		Fields:      nil,
	}
	_, err := timeseries.Create(ctx, db, nt, now)
	if err != nil {
		return err
	}

	return nil
}

func createItem(ctx context.Context, accountID, entityID string, fields map[string]interface{}, db *sqlx.DB, sdb *database.SecDB, fbSDKPath string) error {
	name := "System Generated"
	userID := user.UUID_ENGINE_USER //system user stops workflow from executing hence engine user
	ni := item.NewItem{
		ID:        uuid.New().String(),
		Name:      &name,
		AccountID: accountID,
		EntityID:  entityID,
		UserID:    &userID,
		Fields:    fields,
		Source:    nil,
		Type:      item.TypeDefault,
	}

	it, err := item.Create(ctx, db, ni, time.Now())
	if err != nil {
		return err
	}

	go job.NewJob(db, sdb, fbSDKPath).Stream(stream.NewCreteItemMessage(ctx, db, accountID, userID, entityID, it.ID, ni.Source))
	return nil
}

func findDuplicate(ctx context.Context, accountID string, entityID, uniqueValue interface{}, db *sqlx.DB, sdb *database.SecDB) (*string, error) {
	if uniqueValue == nil || uniqueValue == "" {
		return nil, nil
	} else if entityID == nil || entityID == "" {
		return nil, nil
	}

	conditionFields := make([]graphdb.Field, 0)
	gf := graphdb.Field{
		Expression: "=",
		Key:        "unique",
		DataType:   graphdb.DType(graphdb.TypeString),
		Value:      uniqueValue,
	}
	conditionFields = append(conditionFields, gf)

	useDB := account.UseDB(ctx, db, accountID)
	items, _, err := dbservice.NewDBservice(useDB, db, sdb).Result(ctx, accountID, entityID.(string), "", "", 0, false, false, conditionFields)
	if err != nil {
		return nil, err
	}
	if len(items) > 0 {
		return &items[0].ID, nil
	} else {
		return nil, nil
	}

}

func keyMap(namedKeys map[string]string, namedVals map[string]interface{}) map[string]interface{} {
	itemVals := make(map[string]interface{}, 0)
	for name, key := range namedKeys {
		itemVals[key] = namedVals[name]
	}
	return itemVals
}
