package listeners

import (
	"log"
	"net/http"
	"os"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/mid"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
)

type Listener struct {
	db    *sqlx.DB
	rPool *redis.Pool
}

type Path struct {
	FirebaseSDKPath string
}

// API constructs an http.Handler with all application routes defined.
func API(shutdown chan os.Signal, log *log.Logger, db *sqlx.DB, redisPool *redis.Pool, path Path) http.Handler {

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(shutdown, log, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	workerD := Worker{
		db:                   db,
		rPool:                redisPool,
		firebaseAdminSDKPath: path.FirebaseSDKPath,
	}
	app.Handle("POST", "/v1/sqs/receiver", workerD.receiveSQSPayload)
	// Register health check endpoint. This route is not authenticated.
	check := Check{
		db: db,
	}
	app.Handle("GET", "/v1/health", check.Health)

	return app
}
