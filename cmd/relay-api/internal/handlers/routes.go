package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/mid"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/pubsub"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
)

// API constructs an http.Handler with all application routes defined.
func API(shutdown chan os.Signal, log *log.Logger, db *sqlx.DB, redisPool *redis.Pool, authenticator *auth.Authenticator, publisher *pubsub.Publisher) http.Handler {

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(shutdown, log, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	// Register health check endpoint. This route is not authenticated.
	check := Check{
		db: db,
	}
	app.Handle("GET", "/v1/health", check.Health)

	// Register user management and authentication endpoints.
	u := User{
		db:            db,
		authenticator: authenticator,
	}
	// This route is not authenticated
	app.Handle("GET", "/v1/users/token/:id", u.Token)
	app.Handle("GET", "/v1/users", u.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("POST", "/v1/users", u.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	//this route is confusing?
	app.Handle("GET", "/v1/accounts/users/current/profile", u.Retrieve, mid.Authenticate(authenticator))
	app.Handle("PUT", "/v1/users/:id", u.Update, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("DELETE", "/v1/users/:id", u.Delete, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))

	a := Account{
		db:            db,
		authenticator: authenticator,
	}
	// Register accounts management endpoints.
	app.Handle("GET", "/v1/accounts", a.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser))
	// app.Handle("POST", "/v1/accounts", a.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))

	integ := Integration{
		db:            db,
		authenticator: authenticator,
		publisher:     publisher,
	}
	app.Handle("POST", "/receive/gmail/message", integ.ReceiveEmail)
	app.Handle("GET", "/v1/accounts/:account_id/integrations/:integration_id", integ.AccessIntegration, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("POST", "/v1/accounts/:account_id/integrations/:integration_id", integ.SaveIntegration, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("POST", "/v1/accounts/:account_id/integrations/:integration_id/watch", integ.DailyWatch, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))

	t := Team{
		db:            db,
		authenticator: authenticator,
	}
	// Register teams management endpoints.
	app.Handle("POST", "/v1/accounts/:account_id/teams", t.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams", t.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))

	e := Entity{
		db:            db,
		rPool:         redisPool,
		authenticator: authenticator,
	}
	// Register entities management endpoints.
	// TODO Add team authorization middleware
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/entities", e.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities", e.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id", e.Retrieve, mid.Authenticate(authenticator))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id", e.Update, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))

	i := Item{
		db:            db,
		rPool:         redisPool,
		authenticator: authenticator,
	}
	// Register items management endpoints.
	// TODO Add team authorization middleware
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items", i.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id", i.Update, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items", i.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id", i.Retrieve, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/search", i.Search, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))

	f := Flow{
		db:            db,
		rPool:         redisPool,
		authenticator: authenticator,
	}
	// Register items management endpoints.
	// TODO Add team authorization middleware
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows", f.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows", f.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id", f.Retrieve, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id/items", f.RetrieveActivedItems, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id/items/:item_id", f.RetrieveActiveNodes, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))

	rs := Relationship{
		db:            db,
		authenticator: authenticator,
	}
	// Register relationship management endpoints.
	// TODO Add team authorization middleware
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id/relationships/:relationship_id", rs.ChildItems, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))

	ass := AwsSnsSubscription{
		db: db,
	}
	// Register sns subscription from aws for the product key.
	app.Handle("POST", "/aws/sns/:accountkey/:productkey", ass.Create)

	return app
}
