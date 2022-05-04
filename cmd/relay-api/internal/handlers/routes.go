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

	workerD := Worker{
		db:            db,
		authenticator: authenticator,
		rPool:         redisPool,
	}
	app.Handle("POST", "/v1/sqs/receiver", workerD.receiveSQSPayload)

	// Register user management and authentication endpoints.
	u := User{
		db:            db,
		authenticator: authenticator,
		rPool:         redisPool,
	}

	// users login token
	app.Handle("GET", "/v1/users/verify", u.Verfiy)
	// users join token
	app.Handle("GET", "/v1/users/join", u.Join)
	// users launch token
	app.Handle("GET", "/v1/users/launch", u.Launch)
	// users creation
	app.Handle("POST", "/v1/users", u.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("GET", "/v1/users", u.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	// users profile
	app.Handle("PUT", "/v1/users/:id", u.Update, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("DELETE", "/v1/users/:id", u.Delete, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("GET", "/v1/accounts/users/current/profile", u.Retrieve, mid.Authenticate(authenticator))
	// users invitation
	app.Handle("POST", "/v1/accounts/:account_id/users/invite", u.Invite, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/users/invite", u.Invite, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))

	a := Account{
		db:            db,
		rPool:         redisPool,
		authenticator: authenticator,
	}
	// Register accounts management endpoints.
	app.Handle("POST", "/v1/accounts/drafts", a.Draft)
	// app.Handle("GET", "/v1/accounts/drafts/:draft_id/identifier/:business_email", a.RetriveDraft)
	app.Handle("POST", "/v1/accounts/launch/:draft_id", a.Launch)

	app.Handle("GET", "/v1/accounts/availability", a.Availability)
	app.Handle("GET", "/v1/accounts", a.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser))
	// app.Handle("POST", "/v1/accounts", a.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))

	integ := Integration{
		db:            db,
		rPool:         redisPool,
		authenticator: authenticator,
		publisher:     publisher,
	}
	app.Handle("POST", "/notifications", integ.Notifications)        //google-calendar sync
	app.Handle("POST", "/receive/gmail/message", integ.ReceiveEmail) //gmail sync
	app.Handle("GET", "/v1/accounts/:account_id/integrations/:integration_id", integ.AccessIntegration, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("POST", "/v1/accounts/:account_id/integrations/:integration_id", integ.SaveIntegration, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("POST", "/v1/accounts/:account_id/integrations/:integration_id/actions/:action_id", integ.Act, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))

	t := Team{
		db:            db,
		authenticator: authenticator,
	}
	// Register teams management endpoints.
	app.Handle("POST", "/v1/accounts/:account_id/teams", t.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams", t.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))

	m := Member{
		db:            db,
		rPool:         redisPool,
		authenticator: authenticator,
	}
	// Register teams management endpoints.
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/members", m.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/members/:member_id", m.Update, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/members", m.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))

	e := Entity{
		db:            db,
		rPool:         redisPool,
		authenticator: authenticator,
	}
	// Register entities management endpoints.
	// TODO Add team authorization middleware
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/entities", e.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities", e.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/home", e.Home, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
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
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items", i.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/records/:state", i.StateRecords, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id", i.Update, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id", i.Retrieve, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("DELETE", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id", i.Delete, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/search", i.Search, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/templates", i.CreateTemplate, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))

	s := Segmentation{
		db:            db,
		rPool:         redisPool,
		authenticator: authenticator,
	}
	// Register items management endpoints.
	// TODO Add team authorization middleware
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/segments", s.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))

	f := Flow{
		db:            db,
		rPool:         redisPool,
		authenticator: authenticator,
	}
	// Register items management endpoints.
	// TODO Add team authorization middleware
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows", f.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id", f.Update, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows", f.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id", f.Retrieve, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id/items", f.RetrieveActivedItems, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id/items/:item_id", f.RetrieveActiveNodes, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))

	n := Node{
		db:            db,
		rPool:         redisPool,
		authenticator: authenticator,
	}
	// Register items management endpoints.
	// TODO Add team authorization middleware
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id/nodes", n.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id/nodes/:node_id", n.Update, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id/nodes/:node_id", n.Retrieve, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id/nodes/map", n.Map, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))

	rs := Relationship{
		db:            db,
		rPool:         redisPool,
		authenticator: authenticator,
	}
	// Register relationship management endpoints.
	// TODO Add team authorization middleware
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id/relationships/:relationship_id", rs.ChildItems, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))

	ass := AwsSnsSubscription{
		db:    db,
		rPool: redisPool,
	}
	// Register sns subscription from aws for the product key.
	app.Handle("POST", "/aws/sns/:accountkey/:productkey", ass.Create)

	//TODO move this as a new service
	cv := Conversation{
		db:    db,
		rPool: redisPool,
		hub:   conversation.NewInstanceHub(),
	}
	cv.Listen()
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id/conversations", cv.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id/socket/auth", cv.SocketPreAuth, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/ws/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id/socket/:token", cv.WebSocketMessage, mid.HasSocketAccess(redisPool))
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id/conversations", cv.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))

	ev := Event{
		db:    db,
		rPool: redisPool,
	}
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id/events", ev.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))

	c := Counter{
		db:    db,
		rPool: redisPool,
	}
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/count/:destination", c.Count, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))

	d := Dashboard{
		db:    db,
		rPool: redisPool,
	}
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/overview", d.Overview, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser), mid.HasAccountAccess(db))

	return app
}
