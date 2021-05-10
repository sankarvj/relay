package entity

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/discovery"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"go.opencensus.io/trace"
)

const (
	FixedEntityOwner       = "owners"
	FixedEntityEmailConfig = "email_config"
	FixedEntityCalendar    = "calendar"
	FixedEntityEmails      = "emails"
)

var (
	// ErrFixedEntityNotFound is used when a fixed entity is requested but does not exist.
	ErrFixedEntityNotFound = errors.New("Predefined entity not found")

	// ErrIntegNotFound is used when a specific integrations is requested but none/more than one exist at a time.
	ErrIntegNotFound = errors.New("Integrations not found for fixed entity")
)

type UpdaterFunc func(ctx context.Context, updatedItem interface{}, db *sqlx.DB) error

// EmailEntity represents structural format of email entity
type EmailEntity struct {
	From    []string `json:"from"`
	To      []string `json:"to"`
	Cc      []string `json:"cc"`
	Bcc     []string `json:"bcc"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
}

// EmailConfigEntity represents structural format of email config entity
type EmailConfigEntity struct {
	Domain string   `json:"domain"`
	APIKey string   `json:"api_key"`
	Email  string   `json:"email"`
	Owner  []string `json:"owner"`
	Common string   `json:"common"`
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
	DelayBy string `json:"delay_by"`
	Repeat  string `json:"repeat"`
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
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Email  string `json:"email"`
	Gtoken string `json:"gtoken"`
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

//ParseFixedEntity creates the entity from the given value added fields
func ParseFixedEntity(valueAddedFields []Field, v interface{}) error {
	jsonbody, err := json.Marshal(namedFieldsMap(valueAddedFields))
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonbody, &v)
	return err
}

func RetrieveFixedEntity(ctx context.Context, db *sqlx.DB, accountID string, preDefinedEntity string) (Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.predefined.RetrieveUserEntity")
	defer span.End()

	var e Entity
	const q = `SELECT * FROM entities WHERE account_id = $1 AND name = $2 LIMIT 1`
	if err := db.GetContext(ctx, &e, q, accountID, preDefinedEntity); err != nil {
		if err == sql.ErrNoRows {
			return Entity{}, ErrFixedEntityNotFound
		}

		return Entity{}, errors.Wrapf(err, "selecting pre-defined entity %q", preDefinedEntity)
	}

	return e, nil
}

func RetriveFixedItemByCategory(ctx context.Context, accountID, entityCategory string, db *sqlx.DB) ([]Field, error) {
	fixedEntity, err := RetrieveFixedEntity(ctx, db, accountID, entityCategory)
	if err != nil {
		return nil, err
	}
	items, err := item.EntityItems(ctx, fixedEntity.ID, db)
	if err != nil {
		return nil, err
	}

	if len(items) != 1 {
		return nil, ErrIntegNotFound
	}
	return fixedEntity.ValueAdd(items[0].Fields()), nil
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

func SaveFixedEntityItem(ctx context.Context, accountID, currentUserID, preDefinedEntity string, discoveryID string, namedValues map[string]interface{}, db *sqlx.DB) error {
	fixedEntity, err := RetrieveFixedEntity(ctx, db, accountID, preDefinedEntity)
	if err != nil {
		return err
	}
	entityFields, err := fixedEntity.Fields()
	if err != nil {
		return err
	}

	//delete the old-integrations if present for the specific user
	err = item.DeleteAllByUser(ctx, db, accountID, fixedEntity.ID, currentUserID)
	if err != nil {
		return err
	}

	ni := item.NewItem{
		ID:        uuid.New().String(),
		AccountID: accountID,
		EntityID:  fixedEntity.ID,
		UserID:    &currentUserID,
		Fields:    itemValMap(entityFields, namedValues),
	}

	it, err := item.Create(ctx, db, ni, time.Now())
	if err != nil {
		return err
	}

	if discoveryID != "" {
		ns := discovery.NewDiscovery{
			ID:        discoveryID,
			Type:      integration.TypeGmail,
			AccountID: accountID,
			EntityID:  fixedEntity.ID,
			ItemID:    it.ID,
		}

		_, err = discovery.Create(ctx, db, ns, time.Now())
		if err != nil {
			return err
		}
	}

	return nil
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
