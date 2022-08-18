package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/mid"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/conversation"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
)

// API constructs an http.Handler with all application routes defined.
func API(shutdown chan os.Signal, log *log.Logger, db *sqlx.DB, redisPool *redis.Pool, authenticator *auth.Authenticator, publisher *conversation.Publisher) http.Handler {

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(shutdown, log, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	// Register health check endpoint. This route is not authenticated.
	check := Check{
		db: db,
	}
	app.Handle("GET", "/v1/health", check.Health)
	event := Event{
		db:            db,
		authenticator: authenticator,
	}
	app.Handle("POST", "/v1/slack/event", event.Create, mid.HasSlackAccess(authenticator))

	return app
}
