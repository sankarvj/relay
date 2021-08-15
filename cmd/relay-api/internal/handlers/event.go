package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/platform/event"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

// Check provides support for orchestration health checks.
type Event struct {
	db    *sqlx.DB
	rPool *redis.Pool
	hub   *event.Hub
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

func (ev *Event) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	//TODO now its blindly accepts all orgin. Give the range once when the angular get deployed.
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("error serving web sockets", err)
		return err
	}

	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return web.NewShutdownError("claims missing from context")
	}

	accountID, entityID, itemID := takeAEI(ctx, params, ev.db)
	room := fmt.Sprintf("%s#%s#%s", accountID, entityID, itemID)
	client := event.NewClient(conn, ev.hub, room, currentUserID)

	go client.WritePump(ev.rPool)
	go client.ReadPump(ev.rPool)

	ev.hub.Register <- client

	log.Printf("A new client joined the hub! %+v", client)
	return nil
}

func (ev *Event) Listen() {
	go ev.hub.Run(ev.rPool)
}
