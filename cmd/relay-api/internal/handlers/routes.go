package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/mid"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/conversation"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
)

// API constructs an http.Handler with all application routes defined.
func API(shutdown chan os.Signal, log *log.Logger, db *sqlx.DB, sdb *database.SecDB, authenticator *auth.Authenticator, publisher *conversation.Publisher) http.Handler {

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
		sdb:           sdb,
		authenticator: authenticator,
	}

	// users login token
	app.Handle("GET", "/v1/users/verify", u.Verfiy)
	// join token
	app.Handle("GET", "/v1/users/join", u.Join)
	// visitors token
	app.Handle("GET", "/v1/users/visit", u.Visit)
	// users creation
	app.Handle("POST", "/v1/users", u.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("GET", "/v1/users", u.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	// users profile
	app.Handle("PUT", "/v1/accounts/users/current/profile", u.Update, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("DELETE", "/v1/accounts/users/current/profile", u.Delete, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("GET", "/v1/accounts/users/current/profile", u.Retrieve, mid.Authenticate(authenticator))
	app.Handle("PUT", "/v1/accounts/:account_id/users/current/setting", u.UpdateUserSetting, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember))
	app.Handle("GET", "/v1/accounts/:account_id/users/current/setting", u.RetriveUserSetting, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember))

	a := Account{
		db:            db,
		sdb:           sdb,
		authenticator: authenticator,
	}
	// Register accounts management endpoints.
	app.Handle("POST", "/v1/accounts/drafts", a.Draft)
	app.Handle("POST", "/v1/accounts/launch/:draft_id", a.Launch)

	app.Handle("GET", "/v1/accounts/availability", a.Availability)
	app.Handle("GET", "/v1/accounts", a.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember))
	app.Handle("GET", "/v1/accounts/:account_id", a.Retrieve, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember, auth.RoleUser))
	// app.Handle("POST", "/v1/accounts", a.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))

	v := Visitor{
		db:            db,
		sdb:           sdb,
		authenticator: authenticator,
	}
	// Register accounts management endpoints.
	app.Handle("GET", "/v1/accounts/:account_id/visitors", v.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember))
	app.Handle("GET", "/v1/accounts/:account_id/visitors/:visitor_id", v.Retrieve, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember))
	app.Handle("POST", "/v1/accounts/:account_id/visitors", v.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember))
	app.Handle("PUT", "/v1/accounts/:account_id/visitors/:visitor_id/toggle_active", v.ToggleActive, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("PUT", "/v1/accounts/:account_id/visitors/:visitor_id/resend", v.Resend, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("DELETE", "/v1/accounts/:account_id/visitors/:visitor_id", v.Delete, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))

	integ := Integration{
		db:            db,
		sdb:           sdb,
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
		sdb:           sdb,
		authenticator: authenticator,
	}
	// Register teams management endpoints.
	app.Handle("POST", "/v1/accounts/:account_id/teams", t.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams", t.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("POST", "/v1/accounts/:account_id/teams/templates", t.AddTemplates, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/templates", t.Templates, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/modules", t.Modules, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))

	m := Member{
		db:            db,
		sdb:           sdb,
		authenticator: authenticator,
	}
	// Register teams management endpoints.
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/members", m.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/members/:member_id", m.Update, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/members", m.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("DELETE", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/members/:member_id", m.Delete, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin), mid.HasAccountAccess(db))

	e := Entity{
		db:            db,
		sdb:           sdb,
		authenticator: authenticator,
	}
	// Register entities management endpoints.
	// TODO Add team authorization middleware
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/entities", e.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities", e.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/home", e.Home, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id", e.Retrieve, mid.Authenticate(authenticator))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id", e.Update, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/meta/:ls", e.UpdateLS, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/share", e.ShareTeam, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin), mid.HasAccountAccess(db))
	app.Handle("DELETE", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/share", e.RemoveTeam, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/associate", e.Associations, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleUser))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/associate", e.Associate, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin), mid.HasAccountAccess(db))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/mark", e.Mark, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin), mid.HasAccountAccess(db))
	app.Handle("DELETE", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id", e.Delete, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/toggleaccess", e.ToggleAccess, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin), mid.HasAccountAccess(db))

	fom := Form{
		db:            db,
		sdb:           sdb,
		authenticator: authenticator,
	}
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/forms/:item_id", fom.Render)
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/forms/:item_id", fom.Adder)
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/forms", fom.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))

	noti := Notification{
		db:            db,
		sdb:           sdb,
		authenticator: authenticator,
	}
	// Register teams management endpoints.
	app.Handle("POST", "/v1/accounts/:account_id/notifications/registration", noti.Register, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id/notifications", noti.Clear, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))

	i := Item{
		db:            db,
		sdb:           sdb,
		authenticator: authenticator,
	}
	// Register items management endpoints.
	// TODO Add team authorization middleware
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items", i.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember, auth.RoleUser, auth.RoleVisitor), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items", i.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/records/:state", i.StateRecords, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id", i.Update, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember, auth.RoleUser, auth.RoleMyself), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id", i.Retrieve, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember, auth.RoleUser, auth.RoleVisitor), mid.HasAccountAccess(db))
	app.Handle("DELETE", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id", i.Delete, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember, auth.RoleUser, auth.RoleMyself), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/search", i.Search, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/templates", i.CreateTemplate, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id/toggleaccess", i.ToggleAccess, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin), mid.HasAccountAccess(db))

	s := Segmentation{
		db:            db,
		authenticator: authenticator,
	}
	// Register items management endpoints.
	// TODO Add team authorization middleware
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/segments", s.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember, auth.RoleUser), mid.HasAccountAccess(db))

	f := Flow{
		db:            db,
		sdb:           sdb,
		authenticator: authenticator,
	}
	// Register items management endpoints.
	// TODO Add team authorization middleware
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows", f.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id", f.Update, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id/status", f.UpdateStatus, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows", f.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id", f.Retrieve, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("DELETE", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id", f.Delete, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id/items", f.FlowTrails, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id/items/:item_id", f.TrailNodes, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))

	n := Node{
		db:            db,
		authenticator: authenticator,
	}
	// Register items management endpoints.
	// TODO Add team authorization middleware
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id/nodes", n.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id/nodes/:node_id", n.Update, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id/nodes/:node_id", n.Retrieve, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("DELETE", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id/nodes/:node_id", n.Delete, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/flows/:flow_id/nodes/map", n.Map, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))

	rs := Relationship{
		db:            db,
		sdb:           sdb,
		authenticator: authenticator,
	}
	// Register relationship management endpoints.
	// TODO Add team authorization middleware
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id/relationships/:relationship_id", rs.ChildItems, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember, auth.RoleUser, auth.RoleVisitor), mid.HasAccountAccess(db))

	ass := AwsSnsSubscription{
		db:            db,
		sdb:           sdb,
		authenticator: authenticator,
	}
	// Register sns subscription from aws for the product key.
	app.Handle("POST", "/aws/sns/:accountkey/:productkey", ass.Create)

	//TODO move this as a new service
	cv := Conversation{
		db:            db,
		sdb:           sdb,
		hub:           conversation.NewInstanceHub(),
		authenticator: authenticator,
	}
	cv.Listen()
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id/conversations", cv.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember, auth.RoleUser), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id/socket/auth", cv.SocketPreAuth, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember, auth.RoleUser, auth.RoleVisitor), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/ws/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id/socket/:token", cv.WebSocketMessage, mid.HasSocketAccess(sdb))
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id/conversations", cv.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember, auth.RoleUser), mid.HasAccountAccess(db))

	// Register events endpoints.
	ev := Event{
		db: db,
	}

	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/items/:item_id/events", ev.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))

	// Register counter endpoints.
	c := Counter{
		db:  db,
		sdb: sdb,
	}
	app.Handle("PUT", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/count/:destination", c.Count, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))

	// Register dashboard endpoints.
	d := Dashboard{
		db:  db,
		sdb: sdb,
	}
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/overview", d.Overview, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))

	// Register timeseries endpoints.
	ts := Timeseries{
		db:            db,
		sdb:           sdb,
		authenticator: authenticator,
	}
	app.Handle("POST", "/v1/accounts/:account_id/teams/:team_id/timeseries", ts.Create)
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/timeseries", ts.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/timeseries/:chart_id", ts.Chart, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))
	app.Handle("GET", "/v1/accounts/:account_id/teams/:team_id/entities/:entity_id/timeseries/onme", ts.OnMe, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin, auth.RoleMember), mid.HasAccountAccess(db))

	// Register bill endpoints.
	bi := Bill{
		db:            db,
		authenticator: authenticator,
	}
	app.Handle("POST", "/stripe/webhook", bi.Events)
	app.Handle("POST", "/v1/accounts/:account_id/billing/portal", bi.Portal, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin), mid.HasAccountAccess(db))

	return app
}
