package entity

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/discovery"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific entity is requested but does not exist.
	ErrNotFound = errors.New("Fixed item not found")
)

const (
	DefaultTeamID = "00000000-0000-0000-0000-000000000000"
)

const (
	FixedEntityContacts     = "contacts"
	FixedEntityCompanies    = "companies"
	FixedEntityOwner        = "owners"
	FixedEntityEmailConfig  = "email_config"
	FixedEntityCalendar     = "calendar"
	FixedEntityEmails       = "emails"
	FixedEntityStream       = "stream"
	FixedEntityNotification = "notification"
	//not fixed yet known entities
	FixedEntityTask            = "tasks"
	FixedEntityNote            = "notes"
	FixedEntityMeetings        = "meetings"
	FixedEntityTickets         = "tickets"
	FixedEntityDeals           = "deals"
	FixedEntityProjects        = "projects"
	FixedEntityFlow            = "flows"
	FixedEntityNode            = "nodes"
	FixedEntityDelay           = "delay"
	FixedEntityStatus          = "status"
	FixedEntityType            = "type"
	FixedEntityVisitorInvite   = "visitor_invite"
	FixedEntityEmployee        = "employee"
	FixedEntityPayroll         = "payroll"
	FixedEntitySalary          = "salary"
	FixedEntityAssets          = "assets"
	FixedEntityAssetCatagory   = "asset_catagory"
	FixedEntityAssetRequest    = "asset_request"
	FixedEntityRoles           = "roles"
	FixedEntityServices        = "services"
	FixedEntityServiceCatagory = "service_catagory"
	FixedEntityServiceRequest  = "service_request"
	FixedEntityAgileTask       = "agile_task"
	FixedEntityAgileSubTask    = "agile_sub_task"
)

var (
	// ErrFixedEntityNotFound is used when a fixed entity is requested but does not exist.
	ErrFixedEntityNotFound = errors.New("Predefined entity not found")

	// ErrIntegNotFound is used when a specific integrations is requested but none/more than one exist at a time.
	ErrIntegNotFound = errors.New("Integrations not found for fixed entity")

	ErrIntegAlreadyExists = errors.New("Cannot add this integration. Integrations already exists for that user.")
)

type UpdaterFunc func(ctx context.Context, updatedItem interface{}, db *sqlx.DB) error

// EmailEntity represents structural format of email entity
type EmailEntity struct {
	MessageID   string   `json:"message_id"`
	MessageSent string   `json:"message_sent"`
	From        []string `json:"from"`
	RFrom       []string `json:"rfrom"`
	To          []string `json:"to"`
	Cc          []string `json:"cc"`
	Bcc         []string `json:"bcc"`
	Subject     string   `json:"subject"`
	Body        string   `json:"body"`
	// Contacts    []string `json:"contacts"`
	// Companies   []string `json:"companies"`
}

// EmailConfigEntity represents structural format of email config entity
type EmailConfigEntity struct {
	AccountID string   `json:"account_id"`
	TeamID    string   `json:"team_id"`
	Domain    string   `json:"domain"`
	APIKey    string   `json:"api_key"`
	Email     string   `json:"email"`
	Owner     []string `json:"owner"`
	Common    string   `json:"common"`
	HistoryID string   `json:"history_id"`
}

// CalendarxEntity represents structural format of calendar entity
type CaldendarEntity struct {
	ID        string    `json:"id"`
	APIKey    string    `json:"api_key"`
	Email     string    `json:"email"`
	Owner     []string  `json:"owner"`
	Common    string    `json:"common"`
	SyncedAt  time.Time `json:"synced_at"`
	SyncToken string    `json:"sync_token"`
	Retries   int       `json:"retries"`
}

//DelayEntity represents the structural format of delay entity
type DelayEntity struct {
	Title   string `json:"title"`
	DelayBy int    `json:"delay_by"` // in mins
}

// WebHookEntity represents structural format of webhook entity
type WebHookEntity struct {
	Path    string            `json:"path"`
	Host    string            `json:"host"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
}

// UserEntity represents structural format of user entity
type UserEntity struct {
	MemberID string `json:"member_id"`
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email"`
	Gtoken   string `json:"gtoken"`
}

// FlowEntity represents structural format of flow entity
type FlowEntity struct {
	FlowID      string `json:"flow_id"`
	AccountID   string `json:"account_id"`
	EntityID    string `json:"entity_id"`
	Name        string `json:"name"`
	Expression  string `json:"expression"`
	Description string `json:"description"`
	Mode        int    `json:"mode"`
	Type        int    `json:"type"`
	Condition   int    `json:"condition"`
	Status      int    `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// NotificationEntityItem represents structural format of notification entity
type NotificationEntityItem struct {
	AccountID  string   `json:"account_id"`
	TeamID     string   `json:"team_id"`
	EntityID   string   `json:"entity_id"`
	UserID     string   `json:"user_id"`
	UserName   string   `json:"user_name"`
	UserAvatar string   `json:"user_avatar"`
	ItemID     string   `json:"item_id"`
	Subject    string   `json:"subject"`
	Body       string   `json:"body"`
	Followers  []string `json:"followers"`
	Assignees  []string `json:"assignees"`
	BaseIds    []string `json:"base_ids"`
	Type       int      `json:"type"`
	CreatedAt  string   `json:"created_at"`
}

//ParseFixedEntity creates the entity from the given value added fields
func ParseFixedEntity(valueAddedFields []Field, v interface{}) error {
	jsonbody, err := json.Marshal(namedFieldsMap(valueAddedFields))
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonbody, &v)
	return err
}

func MakeJSONBody(valueAddedFields []Field) ([]byte, error) {
	jsonbody, err := json.Marshal(namedFieldsMap(valueAddedFields))
	if err != nil {
		return nil, err
	}
	return jsonbody, nil
}

func RetrieveFixedEntityAccountLevel(ctx context.Context, db *sqlx.DB, accountID string, preDefinedEntityName string) (Entity, error) {
	ctx, span := trace.StartSpan(ctx, fmt.Sprintf("internal.predefined.RetrieveFixedEntity %s", preDefinedEntityName))
	defer span.End()

	var e Entity
	const q = `SELECT * FROM entities WHERE account_id = $1 AND name = $2 LIMIT 1`
	if err := db.GetContext(ctx, &e, q, accountID, preDefinedEntityName); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("*********> debug internal.entity.fixed entity not found. %s", preDefinedEntityName)
			return Entity{}, ErrFixedEntityNotFound
		}
		return Entity{}, errors.Wrapf(err, "selecting pre-defined  entity %q in account level ", preDefinedEntityName)
	}

	return e, nil
}

func RetrieveFixedEntity(ctx context.Context, db *sqlx.DB, accountID, teamID string, preDefinedEntityName string) (Entity, error) {
	ctx, span := trace.StartSpan(ctx, fmt.Sprintf("internal.predefined.RetrieveFixedEntity %s", preDefinedEntityName))
	defer span.End()

	if teamID == "" {
		teamID = DefaultTeamID
	}

	var e Entity
	const q = "SELECT * FROM entities WHERE account_id = $1 AND name = $2 AND (account_id = $3 OR team_id = $3 OR state = $4 OR shared_team_ids @> $5) LIMIT 1"

	if err := db.GetContext(ctx, &e, q, accountID, preDefinedEntityName, teamID, StateAccountLevel, pq.Array([]string{teamID})); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("*********> debug internal.entity.fixed entity not found. %s", preDefinedEntityName)
			return Entity{}, ErrFixedEntityNotFound
		}
		return Entity{}, errors.Wrapf(err, "selecting pre-defined entity %q", preDefinedEntityName)
	}

	return e, nil
}

func RetriveFixedItemByCategory(ctx context.Context, accountID, teamID, entityCategory string, db *sqlx.DB) ([]Field, UpdaterFunc, error) {
	fixedEntity, err := RetrieveFixedEntity(ctx, db, accountID, teamID, entityCategory)
	if err != nil {
		return nil, nil, err
	}
	items, err := item.EntityItems(ctx, accountID, fixedEntity.ID, db)
	if err != nil {
		return nil, nil, err
	}

	if len(items) != 1 {
		return nil, nil, ErrIntegNotFound
	}
	it := items[0]
	entityFields := fixedEntity.ValueAdd(items[0].Fields())
	return entityFields, updateFields(accountID, fixedEntity.ID, it.ID, entityFields), nil
}

func RetrieveFixedItem(ctx context.Context, accountID, preDefinedEntityID, itemID string, db *sqlx.DB) ([]Field, UpdaterFunc, error) {
	preDefinedEntity, err := Retrieve(ctx, accountID, preDefinedEntityID, db)
	if err != nil {
		return nil, nil, err
	}

	it, err := item.Retrieve(ctx, preDefinedEntityID, itemID, db)
	if err != nil {
		return nil, nil, err
	}

	entityFields := preDefinedEntity.ValueAdd(it.Fields())

	return entityFields, updateFields(accountID, preDefinedEntity.ID, it.ID, entityFields), err
}

func RetrieveUnmarshalledItem(ctx context.Context, accountID, preDefinedEntityID, itemID string, marshalItem interface{}, db *sqlx.DB) (UpdaterFunc, error) {
	valueAddedConfigFields, upFunc, err := RetrieveFixedItem(ctx, accountID, preDefinedEntityID, itemID, db)
	if err != nil {
		return upFunc, err
	}

	err = ParseFixedEntity(valueAddedConfigFields, marshalItem)
	if err != nil {
		return upFunc, err
	}
	return upFunc, nil
}

func SaveFixedEntityItem(ctx context.Context, accountID, teamID, currentUserID, preDefinedEntity, name string, discoveryID, discoveryType string, namedValues map[string]interface{}, db *sqlx.DB) (item.Item, error) {
	fixedEntity, err := RetrieveFixedEntity(ctx, db, accountID, teamID, preDefinedEntity)
	if err != nil {
		return item.Item{}, err
	}
	entityFields, err := fixedEntity.Fields()
	if err != nil {
		return item.Item{}, err
	}

	ni := item.NewItem{
		ID:        uuid.New().String(),
		Name:      &name,
		AccountID: accountID,
		EntityID:  fixedEntity.ID,
		UserID:    &currentUserID,
		Fields:    itemValMap(entityFields, namedValues),
	}

	//check for existence
	if discoveryID != "" {
		dis, err := discovery.Retrieve(ctx, accountID, fixedEntity.ID, discoveryID, db)
		if err != nil && err != discovery.ErrDiscoveryEmpty {
			return item.Item{}, err
		}

		if dis != nil {
			if dis.Type == discoveryType {
				it, err := item.Retrieve(ctx, dis.EntityID, dis.ItemID, db)
				if err != nil {
					return item.Item{}, err
				}
				if *it.UserID == currentUserID { //in some cases we might have to check account level.
					return item.Item{}, ErrIntegAlreadyExists
				}
			}
		}
	}

	it, err := item.Create(ctx, db, ni, time.Now())
	if err != nil {
		return item.Item{}, err
	}

	if discoveryID != "" {
		ns := discovery.NewDiscovery{
			ID:        discoveryID,
			Type:      discoveryType,
			AccountID: accountID,
			EntityID:  fixedEntity.ID,
			ItemID:    it.ID,
		}

		_, err = discovery.Create(ctx, db, ns, time.Now())
		if err != nil {
			return item.Item{}, err
		}
		log.Printf("Discovery item created %+v \n", ns)
	}

	return it, nil
}

func DiscoverAnyEntityItem(ctx context.Context, accountID, entityID, discoveryID string, anyEntityItem interface{}, db *sqlx.DB) (string, error) {
	dis, err := discovery.Retrieve(ctx, accountID, entityID, discoveryID, db)
	if err != nil {
		return "", err
	}
	valueAddedConfigFields, _, err := RetrieveFixedItem(ctx, dis.AccountID, dis.EntityID, dis.ItemID, db)
	if err != nil {
		return "", err
	}

	return dis.ItemID, ParseFixedEntity(valueAddedConfigFields, anyEntityItem)
}

func DiscoverDoneStatusID(ctx context.Context, accountID, entityID string, db *sqlx.DB) (string, error) {
	statusEntity, err := Retrieve(ctx, accountID, entityID, db)
	if err != nil {
		return "", err
	}

	refItems, _ := item.EntityItems(ctx, accountID, statusEntity.ID, db)
	for _, i := range refItems {
		statusFields := statusEntity.ValueAdd(i.Fields())
		for _, statusField := range statusFields {
			if statusField.Name == Verb && statusField.Value == FuExpDone {
				return i.ID, nil
			}
		}
	}
	return "", ErrNotFound
}

//updateFields func encloses the update func
func updateFields(accountID, entityID, itemID string, fields []Field) UpdaterFunc {
	return func(ctx context.Context, updatedItem interface{}, db *sqlx.DB) error {
		_, err := item.UpdateFields(ctx, db, entityID, itemID, itemValMap(fields, util.ConvertInterfaceToMap(updatedItem)))
		return err
	}
}

//itemValMap make the key:value map for storing from the entity and name:value map
func itemValMap(fields []Field, namedFieldsMap map[string]interface{}) map[string]interface{} {
	params := map[string]interface{}{}
	for _, field := range fields {
		params[field.Key] = namedFieldsMap[field.Name]
	}
	return params
}

//namedFieldsMap make the name:value map from the value added entites for fixed entities
func namedFieldsMap(entityFields []Field) map[string]interface{} {
	params := map[string]interface{}{}
	for _, field := range entityFields {
		params[field.Name] = field.Value
	}
	return params
}

func RetriveUserItem(ctx context.Context, accountID, memberID string, db *sqlx.DB) (*UserEntity, error) {
	if memberID == "" {
		return nil, errors.New("memberID is empty. Cannot retrive user item")
	}

	ownerEntity, err := RetrieveFixedEntity(ctx, db, accountID, "", FixedEntityOwner)
	if err != nil {
		return nil, err
	}

	var userEntityItem UserEntity
	valueAddedFields, _, err := RetrieveFixedItem(ctx, ownerEntity.AccountID, ownerEntity.ID, memberID, db)
	if err != nil {
		return nil, err
	}
	err = ParseFixedEntity(valueAddedFields, &userEntityItem)
	if err != nil {
		return nil, err
	}
	userEntityItem.MemberID = memberID
	return &userEntityItem, nil
}
