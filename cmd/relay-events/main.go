package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
)

var (
	eHandler *EventsHandler
)

//informations exists in the handlers only useful for the local testing.
type EventsHandler struct {
	db            *sqlx.DB
	sdb           *database.SecDB
	authenticator *auth.Authenticator
}

func main() {
	eHandler = &EventsHandler{}
	err := initialize(eHandler)
	if err != nil {
		log.Println(errors.Wrap(err, "initializarion error"))
	}
	lambda.Start(eHandler.handleEvent)
}

func initialize(eHandler *EventsHandler) error {

	log := log.New(os.Stdout, "RELAY EVENTS : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	log.Println("main : initialization started")
	// =========================================================================
	// Configuration

	var cfg struct {
		DB struct {
			User       string `conf:"default:postgres,env:DB_USER"`
			Password   string `conf:"default:postgres,noprint,env:DB_PASSWORD"`
			Host       string `conf:"default:0.0.0.0,env:DB_HOST"`
			Name       string `conf:"default:relaydb,env:DB_NAME"`
			DisableTLS bool   `conf:"default:true"`
		}
		Auth struct {
			KeyID              string `conf:"default:1"`
			PrivateKeyFile     string `conf:"default:private.pem,env:AUTH_PRIVATE_KEY_FILE"`
			Algorithm          string `conf:"default:RS256"`
			GoogleKeyFile      string `conf:"default:config/dev/relay-70013-firebase-adminsdk-cfun3-58caec85f0.json,env:AUTH_GOOGLE_KEY_FILE"`
			GoogleClientSecret string `conf:"default:config/dev/google-apps-client-secret.json,env:AUTH_GOOGLE_CLIENT_SECRET"`
		}
		Build string `conf:"default:dev,env:BUILD"`
	}

	if err := conf.Parse(os.Args[1:], "EVENT", &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage("EVENT", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating usage")
			}
			fmt.Println(usage)
			return nil
		}
		return errors.Wrap(err, "error: parsing config")
	}

	dbConfig := database.Config{
		User:       cfg.DB.User,
		Password:   cfg.DB.Password,
		Host:       cfg.DB.Host,
		Name:       cfg.DB.Name,
		DisableTLS: cfg.DB.DisableTLS,
	}

	//connect to db only if connection is nil
	var err error
	log.Println("main : started : initializing database")
	if eHandler.db == nil {
		eHandler.db, err = database.Open(dbConfig)
		if err != nil {
			return errors.Wrap(err, "connecting to primary db")
		}
		log.Println("main : completed : new db instance created")
	} else {
		log.Println("main : completed : old db instance reused")
	}

	log.Println("main : started : initializing authentication support")
	err = initializeAuth(eHandler, cfg.Auth.PrivateKeyFile, cfg.Auth.KeyID, cfg.Auth.GoogleKeyFile, cfg.Auth.GoogleClientSecret, cfg.Auth.Algorithm)
	if err != nil {
		return errors.Wrap(err, "main : errored : authentication support")
	}
	log.Println("main : completed : authentication support")
	log.Println("main : initialization completed")

	return nil
}

func initializeAuth(eHandler *EventsHandler, privateKeyFile, keyID, googleKeyFile, secret, algo string) error {
	keyContents, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		return errors.Wrap(err, "reading auth private key")
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyContents)
	if err != nil {
		return errors.Wrap(err, "parsing auth private key")
	}

	f := auth.NewSimpleKeyLookupFunc(keyID, privateKey.Public().(*rsa.PublicKey))
	eHandler.authenticator, err = auth.NewAuthenticator(privateKey, googleKeyFile, secret, keyID, algo, f)
	if err != nil {
		return errors.Wrap(err, "constructing authenticator")
	}
	return nil
}

func (h EventsHandler) handleEvent(ctx context.Context, event interface{}) error {
	payload, ok := event.(map[string]interface{})
	if !ok {
		return web.NewRequestError(errors.New("post body not exist"), http.StatusBadRequest)
	}

	var body map[string]interface{}
	if payload["body"] != nil {
		body = payload["body"].(map[string]interface{})
	} else {
		return web.NewRequestError(errors.New("post body not exist"), http.StatusBadRequest)
	}

	headers := payload["headers"].(map[string]interface{})
	token := headers["Authorization"]

	accountID, err := h.authenticate(token)
	if err != nil {
		return err
	}
	fmt.Println("handleEvent : authentication successfull")

	return h.processEvent(ctx, accountID, body)
}

func (h EventsHandler) processEvent(ctx context.Context, accountID string, body map[string]interface{}) error {
	fmt.Println("processEvent : started")
	identifier := strValue(body["identifier"])
	entityName := strValue(body["module"])

	//actual entity : page_view, events, errors, sign_ups, subscriptions
	e, err := entity.RetrieveByName(ctx, accountID, entityName, h.db)
	if err != nil {
		return err
	}

	items, err := item.List(ctx, accountID, e.ID, h.db)
	if err != nil {
		return err
	}

	it := item.Item{}
	if len(items) > 0 {
		it = items[0]
	}

	itemFields := it.Fields()
	if itemFields == nil {
		itemFields = make(map[string]interface{}, 0)
	}
	namedFieldsMap := e.NamedFields()

	for name, v := range body {
		if f, ok := namedFieldsMap[name]; ok {
			itemFields[f.Key] = f.CalcFunc().Calc(itemFields[f.Key], v)
		}
	}

	if it.ID == "" {
		it, err = createItem(ctx, h.db, accountID, e.ID, itemFields)
		if err != nil {
			return err
		}
		job.NewJob(h.db, h.sdb, h.authenticator.FireBaseAdminSDK).Stream(stream.NewEventItemMessage(ctx, h.db, accountID, "", e.ID, it.ID, it.Fields(), nil, identifier))
	} else {
		updatedItem, err := item.UpdateFields(ctx, h.db, e.ID, it.ID, itemFields)
		if err != nil {
			return err
		}
		job.NewJob(h.db, h.sdb, h.authenticator.FireBaseAdminSDK).Stream(stream.NewEventItemMessage(ctx, h.db, accountID, "", e.ID, it.ID, updatedItem.Fields(), it.Fields(), identifier))
	}

	fmt.Println("processEvent : completed")

	return nil
}

func createItem(ctx context.Context, db *sqlx.DB, accountID, entityID string, fields map[string]interface{}) (item.Item, error) {
	ni := item.NewItem{
		ID:        uuid.New().String(),
		Name:      nil,
		AccountID: accountID,
		EntityID:  entityID,
		UserID:    nil,
		Fields:    fields,
	}
	it, err := item.Create(ctx, db, ni, time.Now())
	if err != nil {
		return it, err
	}
	return it, nil
}

func (h EventsHandler) authenticate(token interface{}) (string, error) {
	if token == nil {
		err := errors.New("token does not exist")
		return "", web.NewRequestError(err, http.StatusUnauthorized)
	}

	tokenStr, ok := token.(string)
	if !ok {
		err := errors.New("token is not valid")
		return "", web.NewRequestError(err, http.StatusUnauthorized)
	}

	parts := strings.Split(tokenStr, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		err := errors.New("expected authorization header format: Bearer <token>")
		return "", web.NewRequestError(err, http.StatusUnauthorized)
	}

	claims, err := h.authenticator.ParseClaims(parts[1])
	if err != nil {
		return "", web.NewRequestError(err, http.StatusUnauthorized)
	}

	return claims.Subject, nil
}

func strValue(v interface{}) string {
	if v != nil {
		return v.(string)
	}
	return ""
}
