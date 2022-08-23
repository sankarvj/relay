package handlers

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/account"
	conv "gitlab.com/vjsideprojects/relay/internal/conversation"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/conversation"
	"gitlab.com/vjsideprojects/relay/internal/platform/redisdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/bcrypt"
)

// Check provides support for orchestration health checks.
type Conversation struct {
	db            *sqlx.DB
	rPool         *redis.Pool
	hub           *conversation.Hub
	authenticator *auth.Authenticator
	MessageChan   chan conversation.Message // to receive message in the handler from the hub platform
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

// List returns all the existing entities associated with team
func (cv *Conversation) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Conversation.List")
	defer span.End()

	acc, err := account.Retrieve(ctx, cv.db, params["account_id"])
	if err != nil {
		return err
	}

	conversations, err := conv.List(ctx, params["account_id"], params["entity_id"], params["item_id"], 0, cv.db)
	if err != nil {
		return err
	}

	viewModelConversations := make([]conv.ViewModelConversation, len(conversations))
	for i, cnv := range conversations {
		viewModelConversations[i] = createViewModelConversation(cnv, acc.Name)
	}

	return web.Respond(ctx, w, viewModelConversations, http.StatusOK)
}

func (cv *Conversation) SocketPreAuth(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return errors.Wrapf(err, "auth claims missing from context")
	}

	token := generateToken(currentUserID)

	err = redisdb.RedisSet(cv.rPool, token, currentUserID)
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

func (cv *Conversation) WebSocketMessage(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	//TODO now its blindly accepts all orgin. Give the range once when the angular get deployed.
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("***> unexpected error occurred when serving web sockets. error: %v\n", err)
		return err
	}

	currentUserID, err := user.RetrieveWSCurrentUserID(ctx)
	if err != nil {
		return errors.Wrapf(err, "auth claims missing from context")
	}

	cuser, err := user.RetrieveUser(ctx, cv.db, currentUserID)
	if err != nil {
		return err
	}

	clientID := uuid.New().String()
	accountID, entityID, itemID := takeAEI(ctx, params, cv.db)
	room := fmt.Sprintf("%s#%s#%s", accountID, entityID, itemID)
	client := conversation.NewClient(conn, cv.hub, clientID, room, currentUserID, *cuser.Name, *cuser.Avatar)

	go client.WritePump(cv.rPool)
	go client.ReadPump(cv.rPool, cv.MessageChan)

	cv.hub.Register <- client

	log.Printf("internal.handlers.event new client joined the hub! %+v\n", client)
	return nil
}

func (cv *Conversation) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Conversation.Create")
	defer span.End()

	currentUser, err := user.RetrieveCurrentUser(ctx, cv.db)
	if err != nil {
		return err
	}

	var nc conv.NewConversation
	if err := web.Decode(r, &nc); err != nil {
		return errors.Wrap(err, "")
	}

	nc.ID = uuid.New().String()
	nc.AccountID = params["account_id"]
	nc.EntityID = params["entity_id"]
	itemID := params["item_id"]
	nc.ItemID = &itemID
	nc.Type = conv.TypeConvSent
	nc.UserID = currentUser.ID //TODO store name and avatar also

	conversation, err := conv.Create(ctx, cv.db, nc, time.Now())
	if err != nil {
		return err
	}

	job.NewJob(cv.db, cv.rPool, cv.authenticator.FireBaseAdminSDK).Stream(stream.NewConversationMessage(ctx, cv.db, params["account_id"], currentUser.ID, params["entity_id"], itemID, conversation.ID))

	return web.Respond(ctx, w, conversation, http.StatusCreated)
}

//WS
func (cv *Conversation) Listen() {
	go cv.runGlobalMessageReceiver()
	go cv.hub.Run(cv.rPool)
}

func (cv *Conversation) runGlobalMessageReceiver() {
	cv.MessageChan = make(chan conversation.Message)
	for {
		select {
		case newMessage := <-cv.MessageChan:
			parts := strings.Split(newMessage.Room, "#")
			newConversation := conv.NewConversation{
				ID:        newMessage.Payload.ID,
				AccountID: parts[0],
				EntityID:  parts[1],  // stream entity
				ItemID:    &parts[2], // stream item
				UserID:    newMessage.Payload.UserID,
				Message:   newMessage.Payload.Message,
			}
			_, err := conv.Create(context.Background(), cv.db, newConversation, time.Now())
			if err != nil {
				log.Println("***> unhandled unexpected error occurred. when inserting the chat message to the DB", err)
			}
		}
	}
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

func createViewModelConversation(cnu conv.ConversationUsr, accName string) conv.ViewModelConversation {

	vmc := conv.ViewModelConversation{
		ID:        cnu.ID,
		Message:   cnu.Message,
		Type:      cnu.Type,
		State:     cnu.State,
		CreatedAt: cnu.CreatedAt,
	}

	if cnu.UserID != nil {
		vmc.UserID = *cnu.UserID
	} else {
		vmc.UserID = schema.SeedSystemUserID
	}

	if cnu.UserAvatar != nil {
		vmc.UserAvatar = *cnu.UserAvatar
	} else {
		vmc.UserAvatar = fmt.Sprintf("https://avatars.dicebear.com/api/bottts/%s.svg", accName)
	}

	if cnu.UserName != nil {
		vmc.UserName = *cnu.UserName
	} else {
		vmc.UserName = accName
	}

	return vmc
}
