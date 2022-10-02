package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"expvar"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ardanlabs/conf"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dgrijalva/jwt-go"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/event"
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
	log           *log.Logger
}

func newEventsHandler() *EventsHandler {
	if eHandler == nil { // create if not exists already
		eHandler = &EventsHandler{}
		err := initialize(eHandler)
		if err != nil {
			eHandler.log.Println(errors.Wrap(err, "initializarion error"))
		}
	}
	return eHandler
}

func initialize(eHandler *EventsHandler) error {
	eHandler.log = log.New(os.Stdout, "RELAY EVENTS : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	eHandler.log.Println("main : initialization started")

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
		Service struct {
			Region       string `conf:"default:us-east-1,env:AWS_REGION"`
			WorkerSqsURL string `conf:"default:us-east-1,env:AWS_WORKER_SQS_URL"`
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

	// Store Global Variables
	expvar.NewString("build").Set(cfg.Build)
	expvar.NewString("aws_region").Set(cfg.Service.Region)
	expvar.NewString("aws_worker_sqs_url").Set(cfg.Service.WorkerSqsURL)

	// =========================================================================
	// App Starting
	eHandler.log.Printf("main : started : application initializing : version %q", cfg.Build)
	dbConfig := database.Config{
		User:       cfg.DB.User,
		Password:   cfg.DB.Password,
		Host:       cfg.DB.Host,
		Name:       cfg.DB.Name,
		DisableTLS: cfg.DB.DisableTLS,
	}

	//connect to db only if connection is nil
	var err error
	eHandler.log.Println("main : started : initializing database")
	if eHandler.db == nil {
		eHandler.db, err = database.Open(dbConfig)
		if err != nil {
			return errors.Wrap(err, "connecting to primary db")
		}
		eHandler.log.Println("main : completed : new db instance created")
	}

	eHandler.log.Println("main : started : initializing authentication support")
	err = initializeAuth(eHandler, cfg.Auth.PrivateKeyFile, cfg.Auth.KeyID, cfg.Auth.GoogleKeyFile, cfg.Auth.GoogleClientSecret, cfg.Auth.Algorithm)
	if err != nil {
		return errors.Wrap(err, "main : errored : initializing authentication support")
	}
	eHandler.log.Println("main : completed : initializing authentication support")
	eHandler.log.Println("main : initialization completed")

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

func main() {
	lambda.Start(newEventsHandler().handleEvent)
}

func (h EventsHandler) handleEvent(ctx context.Context, event interface{}) (events.APIGatewayProxyResponse, error) {
	eHandler.log.Println("handleEvent : started : parse payload")
	payload, ok := event.(map[string]interface{})
	if !ok {
		eHandler.log.Println("handleEvent : errored : parse payload")
		return newErrReponse(errors.New("post payload not exist"))
	}

	var body map[string]interface{}
	if payload["body"] != nil {
		jsonStr := payload["body"].(string)
		err := json.Unmarshal([]byte(jsonStr), &body)
		if err != nil {
			eHandler.log.Println("handleEvent : errored : parse body")
			return newErrReponse(errors.New("post body not in proper format"))
		}
	} else {
		eHandler.log.Println("handleEvent : errored : parse body empty")
		return newErrReponse(errors.New("post body not exist"))
	}

	headers := payload["headers"].(map[string]interface{})
	token := headers["authorization"]
	eHandler.log.Println("handleEvent : completed : parse payload")

	eHandler.log.Println("handleEvent : started : authentication")
	accountID, err := h.authenticate(token)
	if err != nil {
		eHandler.log.Println("handleEvent : errored : authentication")
		return newErrReponse(err)
	}
	eHandler.log.Println("handleEvent : completed : authentication")

	return h.processEvent(ctx, accountID, body)
}

func (h EventsHandler) processEvent(ctx context.Context, accountID string, body map[string]interface{}) (events.APIGatewayProxyResponse, error) {
	eHandler.log.Println("processEvent : started : save event")
	identifier := strValue(body["identifier"])
	entityName := strValue(body["module"])

	itEve, err := event.Process(ctx, accountID, entityName, body, eHandler.log, h.db)
	if err != nil {
		return newErrReponse(err)
	}

	if itEve.OldItem == nil && itEve.NewItem != nil {
		eHandler.log.Println("processEvent : started : sqs streaming")
		err = job.NewJob(h.db, h.sdb, h.authenticator.FireBaseAdminSDK).Stream(stream.NewEventItemMessage(ctx, h.db, accountID, "", itEve.NewItem.EntityID, itEve.NewItem.ID, itEve.NewItem.Fields(), nil, identifier))
		if err != nil {
			eHandler.log.Println("processEvent : errored : sqs streaming", err)
		}
		eHandler.log.Println("processEvent : completed : sqs streaming")
	} else if itEve.OldItem != nil && itEve.NewItem != nil {
		eHandler.log.Println("processEvent : started : sqs streaming")
		err = job.NewJob(h.db, h.sdb, h.authenticator.FireBaseAdminSDK).Stream(stream.NewEventItemMessage(ctx, h.db, accountID, "", itEve.NewItem.EntityID, itEve.NewItem.ID, itEve.NewItem.Fields(), itEve.OldItem.Fields(), identifier))
		if err != nil {
			eHandler.log.Println("processEvent : errored : sqs streaming", err)
		}
		eHandler.log.Println("processEvent : completed : sqs streaming")
	}

	return newSuccessReponse("success")
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

func newSuccessReponse(message string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       message,
		StatusCode: 200,
	}, nil
}

func newErrReponse(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       err.Error(),
		StatusCode: 500,
	}, err
}
