package entity

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"go.opencensus.io/trace"
)

const (
	FixedEntityOwner       = "owners"
	FixedEntityEmailConfig = "email_config"
	FixedEntityEmails      = "emails"
)

var (
	// ErrFixedEntityNotFound is used when a fixed entity is requested but does not exist.
	ErrFixedEntityNotFound = errors.New("Predefined entity not found")
)

type updaterFunc func(ctx context.Context, updatedItem interface{}, db *sqlx.DB) error

// EmailEntity represents structural format of email entity
type EmailEntity struct {
	Config  string `json:"config"`
	From    string `json:"from"`
	To      string `json:"to"`
	Cc      string `json:"cc"`
	Bcc     string `json:"bcc"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// EmailConfigEntity represents structural format of email config entity
type EmailConfigEntity struct {
	Domain string `json:"domain"`
	APIKey string `json:"api_key"`
	Email  string `json:"email"`
	Owner  string `json:"owner"`
	Common string `json:"common"`
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

func RetrieveFixedItem(ctx context.Context, accountID, preDefinedEntityID, itemID string, db *sqlx.DB) ([]Field, updaterFunc, error) {
	preDefinedEntity, err := RetrieveFixedEntity(ctx, db, accountID, preDefinedEntityID)
	if err != nil {
		return nil, nil, err
	}

	it, err := item.Retrieve(ctx, preDefinedEntity.ID, itemID, db)
	if err != nil {
		return nil, nil, err
	}

	entityFields, err := preDefinedEntity.Fields()
	if err != nil {
		return nil, nil, err
	}

	entityFields = FillFieldValues(entityFields, it.Fields())

	return entityFields, updateFields(accountID, preDefinedEntity.ID, it.ID, entityFields), err
}

func SaveEmailIntegration(ctx context.Context, accountID, currentUserID, domain, token, emailAddress string, db *sqlx.DB) (item.Item, error) {
	emailConfigEntity, err := RetrieveFixedEntity(ctx, db, accountID, FixedEntityEmailConfig)
	if err != nil {
		return item.Item{}, err
	}

	entityFields, err := emailConfigEntity.AllFields()
	if err != nil {
		return item.Item{}, err
	}

	var emailConfigEntityItem EmailConfigEntity
	err = ParseFixedEntity(entityFields, &emailConfigEntityItem)
	if err != nil {
		return item.Item{}, err
	}
	emailConfigEntityItem.APIKey = token
	emailConfigEntityItem.Domain = domain
	emailConfigEntityItem.Email = emailAddress
	emailConfigEntityItem.Common = "false"
	emailConfigEntityItem.Owner = currentUserID

	ni := item.NewItem{
		ID:        uuid.New().String(),
		AccountID: accountID,
		EntityID:  emailConfigEntity.ID,
		Fields:    itemValMap(entityFields, util.ConvertInterfaceToMap(emailConfigEntityItem)),
	}

	return item.Create(ctx, db, ni, time.Now())

}

func SaveEmailTemplate(ctx context.Context, accountID, emailConfigItemID, to, cc, bcc, subject, body string, db *sqlx.DB) (item.Item, error) {
	emailEntity, err := RetrieveFixedEntity(ctx, db, accountID, FixedEntityEmails)
	if err != nil {
		return item.Item{}, err
	}

	entityFields, err := emailEntity.AllFields()
	if err != nil {
		return item.Item{}, err
	}

	var emailEntityItem EmailEntity
	err = ParseFixedEntity(entityFields, &emailEntityItem)
	if err != nil {
		return item.Item{}, err
	}
	emailEntityItem.Config = emailConfigItemID
	emailEntityItem.From = emailConfigItemID
	emailEntityItem.To = to
	emailEntityItem.Cc = cc
	emailEntityItem.Bcc = bcc
	emailEntityItem.Subject = subject
	emailEntityItem.Body = body

	ni := item.NewItem{
		ID:        uuid.New().String(),
		AccountID: accountID,
		EntityID:  emailEntity.ID,
		Fields:    itemValMap(entityFields, util.ConvertInterfaceToMap(emailEntityItem)),
	}

	return item.Create(ctx, db, ni, time.Now())

}

//updateFields func encloses the update func
func updateFields(accountID, entityID, itemID string, fields []Field) updaterFunc {
	return func(ctx context.Context, updatedItem interface{}, db *sqlx.DB) error {
		return item.UpdateFields(ctx, db, entityID, itemID, itemValMap(fields, util.ConvertInterfaceToMap(updatedItem)))
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
