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
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
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
	log.Println("initialization started")
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
		SecDB struct {
			User     string `conf:"default:redisgraph,env:SEC_DB_USER"`
			Password string `conf:"default:redis,noprint,env:SEC_DB_PASSWORD"`
			Host     string `conf:"default:127.0.0.1:6379,env:SEC_DB_HOST"`
			Name     string `conf:"default:relaydb,env:SEC_DB_NAME"`
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
	if eHandler.db == nil {
		eHandler.db, err = database.Open(dbConfig)
		if err != nil {
			return errors.Wrap(err, "connecting to primary db")
		}
		log.Println("new db instance created")
	} else {
		log.Println("old db instance reused")
	}

	secDbConfig := database.SecConfig{
		User:     cfg.SecDB.User,
		Password: cfg.SecDB.Password,
		Host:     cfg.SecDB.Host,
		Name:     cfg.SecDB.Name,
	}

	if eHandler.sdb == nil || eHandler.sdb.GraphPool() == nil {
		rp := &redis.Pool{
			MaxIdle:     50,
			MaxActive:   50,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", secDbConfig.Host, redis.DialPassword(secDbConfig.Password))
				if err != nil {
					return nil, err
				}
				return c, err
			},

			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		}
		eHandler.sdb = database.Init(rp, rp, rp)
		log.Println("new sdb instance created")
	} else {
		log.Println("old sdb instance reused")
	}

	log.Println("main : Started : Initializing authentication support")
	err = initializeAuth(eHandler, cfg.Auth.PrivateKeyFile, cfg.Auth.KeyID, cfg.Auth.GoogleKeyFile, cfg.Auth.GoogleClientSecret, cfg.Auth.Algorithm)
	if err != nil {
		return errors.Wrap(err, "initializing authentication support")
	}
	log.Println("main : Completed : Initializing authentication support")
	log.Println("initialization completed")

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
	fmt.Printf("received event: %+v\n", event)

	payload, ok := event.(map[string]interface{})
	if !ok {
		err := errors.New("post body does not exist")
		return web.NewRequestError(err, http.StatusBadRequest)
	}

	var body map[string]interface{}
	if payload["body"] != nil {
		body = payload["body"].(map[string]interface{})
	} else {
		err := errors.New("body does not exist")
		return web.NewRequestError(err, http.StatusBadRequest)
	}

	headers := payload["headers"].(map[string]interface{})
	token := headers["Authorization"]
	fmt.Printf("received body: %+v\n", body)

	accountID, err := h.authenticate(token)
	if err != nil {
		return err
	}
	fmt.Println("authentication successfull")

	h.findingBase(ctx, accountID, strValue(body["identifier"]))

	entityName := strValue(body["module"])
	data := make(map[string]interface{}, 0)

	return h.processEvent(ctx, accountID, entityName, data)
}

func (h EventsHandler) processEvent(ctx context.Context, accountID, entityName string, data map[string]interface{}) error {
	log.Println("entityName ", entityName)
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
	for name, v := range data {
		if f, ok := namedFieldsMap[name]; ok {
			if f.Who == entity.WhoContacts {
				itemFields[f.Key] = []interface{}{}
			} else if f.Who == entity.WhoCompanies {
				itemFields[f.Key] = []interface{}{}
			} else {
				itemFields[f.Key] = f.CalcFunc().Calc(itemFields[f.Key], v)
			}
		}
	}

	if it.ID == "" {
		err = createItem(ctx, h.db, accountID, e.ID, itemFields)
		if err != nil {
			return err
		}
		go job.NewJob(h.db, h.sdb, h.authenticator.FireBaseAdminSDK).Stream(stream.NewCreteItemMessage(ctx, h.db, accountID, "", e.ID, it.ID, map[string][]string{}))
	} else {
		updatedItem, err := item.UpdateFields(ctx, h.db, e.ID, it.ID, itemFields)
		if err != nil {
			return err
		}
		go job.NewJob(h.db, h.sdb, h.authenticator.FireBaseAdminSDK).Stream(stream.NewUpdateItemMessage(ctx, h.db, accountID, "", e.ID, it.ID, updatedItem.Fields(), it.Fields()))
	}

	return nil
}

func createItem(ctx context.Context, db *sqlx.DB, accountID, entityID string, fields map[string]interface{}) error {
	ni := item.NewItem{
		ID:        uuid.New().String(),
		Name:      nil,
		AccountID: accountID,
		EntityID:  entityID,
		UserID:    nil,
		Fields:    fields,
	}
	_, err := item.Create(ctx, db, ni, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (h EventsHandler) findingBase(ctx context.Context, accountID string, identifier string) (map[string]string, error) {
	sourceMap := map[string]string{}
	elements := strings.Split(identifier, ":")
	if len(elements) == 3 {
		conditionFields := make([]graphdb.Field, 0)
		entityName := elements[0]
		fieldKey := elements[1]
		fieldValue := elements[1]

		e, err := entity.RetrieveByName(ctx, accountID, entityName, h.db)
		if err != nil {
			return sourceMap, err
		}
		exp := fmt.Sprintf("{{%s.%s}} eq {%s}", e.ID, fieldKey, fieldValue)
		filter := job.NewJabEngine().RunExpGrapher(ctx, h.db, h.sdb, accountID, exp)

		fields, err := e.FilteredFields()
		if err != nil {
			return sourceMap, err
		}

		for _, f := range fields {
			if condition, ok := filter.Conditions[f.Key]; ok {
				conditionFields = append(conditionFields, f.MakeGraphField(condition.Term, condition.Expression, false))
			}
		}

		gSegment := graphdb.BuildGNode(accountID, e.ID, false).MakeBaseGNode("", conditionFields)
		result, err := graphdb.GetResult(h.sdb.GraphPool(), gSegment, 0, "", "")
		log.Println("three result", result)
	}
	return sourceMap, nil
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
