package handlers

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/platform/event"
	"gitlab.com/vjsideprojects/relay/internal/platform/redisdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"golang.org/x/crypto/bcrypt"
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

func (ev *Event) Retrive(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return web.NewShutdownError("auth claims missing from context")
	}

	token := generateToken(currentUserID)

	err = redisdb.RedisSet(ev.rPool, token, currentUserID)
	if err != nil {
		return err
	}

	socketAuth := struct {
		Token string `json:"token"`
	}{
		token,
	}
	return web.Respond(ctx, w, socketAuth, http.StatusCreated)
}

func (ev *Event) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	//TODO now its blindly accepts all orgin. Give the range once when the angular get deployed.
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("unexpected error occurred when serving web sockets. error: %v\n", err)
		return err
	}

	currentUserID, err := user.RetrieveWSCurrentUserID(ctx)
	if err != nil {
		return web.NewShutdownError("auth claims missing from context")
	}

	clientID := uuid.New().String()
	accountID, entityID, itemID := takeAEI(ctx, params, ev.db)
	room := fmt.Sprintf("%s#%s#%s", accountID, entityID, itemID)
	client := event.NewClient(conn, ev.hub, clientID, room, currentUserID)

	go client.WritePump(ev.rPool)
	go client.ReadPump(ev.rPool)

	ev.hub.Register <- client

	log.Printf("internal.handlers.event new client joined the hub! %+v\n", client)
	return nil
}

func (ev *Event) Listen() {
	go ev.hub.Run(ev.rPool)
}

func generateToken(email string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(email), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	hasher := md5.New()
	hasher.Write(hash)
	return hex.EncodeToString(hasher.Sum(nil))
}
