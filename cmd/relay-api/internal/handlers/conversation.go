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

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/account"
	conv "gitlab.com/vjsideprojects/relay/internal/conversation"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/notification"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/conversation"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/bcrypt"
)

// Check provides support for orchestration health checks.
type Conversation struct {
	db            *sqlx.DB
	sdb           *database.SecDB
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

	err = cv.sdb.SetSocketAuthToken(token, currentUserID)
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
	baseEntityID := r.URL.Query().Get("be")
	baseItemID := r.URL.Query().Get("bi")
	base := fmt.Sprintf("%s#%s", baseEntityID, baseItemID)
	room := fmt.Sprintf("%s#%s#%s", accountID, entityID, itemID)
	client := conversation.NewClient(conn, cv.hub, clientID, base, room, currentUserID, cuser.Email, *cuser.Name, *cuser.Avatar)

	go client.WritePump()
	go client.ReadPump(cv.sdb.PubSubPool(), cv.MessageChan)

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

	job.NewJob(cv.db, cv.sdb, cv.authenticator.FireBaseAdminSDK).Stream(stream.NewEmailConversationMessage(ctx, cv.db, params["account_id"], currentUser.ID, params["entity_id"], itemID, conversation.ID, map[string][]string{}))

	return web.Respond(ctx, w, conversation, http.StatusCreated)
}

//WS
func (cv *Conversation) Listen() {
	go cv.runGlobalMessageReceiver()
	go cv.hub.Run(cv.sdb.PubSubPool())
}

func (cv *Conversation) runGlobalMessageReceiver() {
	cv.MessageChan = make(chan conversation.Message)

	for {
		select {
		case newMessage := <-cv.MessageChan:
			//authenticate before inserting it to the DB
			baseParts := strings.Split(newMessage.Base, "#")
			streamParts := strings.Split(newMessage.Room, "#")
			newConversation := conv.NewConversation{
				ID:        newMessage.Payload.ID,
				AccountID: streamParts[0],
				EntityID:  streamParts[1],  // stream entity
				ItemID:    &streamParts[2], // stream item
				UserID:    newMessage.Payload.UserID,
				Message:   newMessage.Payload.Message,
			}
			ctx := context.Background()
			_, err := conv.Create(context.Background(), cv.db, newConversation, time.Now())
			if err != nil {
				log.Println("***> unhandled unexpected error occurred. when inserting the chat message to the DB", err)
			}

			//The conv added function works differently
			go job.NewJob(cv.db, cv.sdb, cv.authenticator.FireBaseAdminSDK).Stream(stream.NewChatConversationMessage(ctx, cv.db, newConversation.AccountID, newConversation.UserID, newConversation.EntityID, *newConversation.ItemID, newMessage.Payload.ID, map[string][]string{baseParts[0]: {baseParts[1]}}))
			//addNotification(ctx, streamParts[0], baseParts[0], baseParts[1], newConversation.UserID, newConversation.Message, cv.authenticator.FireBaseAdminSDK, cv.db)
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

func addNotification(ctx context.Context, accountID, entityID, itemID, userID, message string, firebaseSDKPath string, db *sqlx.DB, sdb *database.SecDB) error {
	baseEntity, err := entity.Retrieve(ctx, accountID, entityID, db, sdb)
	if err != nil {
		log.Println("***>***> addNotification: unexpected/unhandled error occurred when retriving entity on job. error:", err)
		return err
	}
	it, err := item.Retrieve(ctx, entityID, itemID, db)
	if err != nil {
		log.Println("***>***> addNotification: unexpected/unhandled error occurred while retriving item on job. error:", err)
		return err
	}

	appNotif := notification.AppNotification{
		AccountID: accountID,
		TeamID:    "", // not able to get the team-id here....
		UserID:    userID,
		EntityID:  entityID,
		ItemID:    itemID,
		Followers: make([]entity.UserEntity, 0),
		Assignees: make([]entity.UserEntity, 0),
		BaseIds:   make([]string, 0), //events filter use case. check README for more info
	}
	titleField := entity.TitleField(baseEntity.EasyFields())
	appNotif.BaseItemName = it.Fields()[titleField.Key].(string)
	//adding base item follower and assignees
	appNotif.AddFollower(ctx, accountID, it.UserID, db, sdb)
	baseValueAddedFields := baseEntity.ValueAdd(it.Fields())
	for _, f := range baseValueAddedFields {
		if f.Value == nil {
			continue
		}
		if f.Who == entity.WhoAssignee {
			appNotif.AddAssignees(ctx, accountID, f.RefID, f.Value.([]interface{}), db, sdb)
		}
	}

	appNotif.BaseIds = append(appNotif.BaseIds, fmt.Sprintf("%s#%s", baseEntity.ID, it.ID))
	appNotif.Subject = fmt.Sprintf("Comment added in %s `%s`", util.LowerSinglarize(baseEntity.DisplayName), appNotif.BaseItemName)
	appNotif.Body = util.TruncateText(message, 20)

	notificationType := notification.TypeChatConversationAdded
	duplicateMasker := make(map[string]bool, 0)
	//Send email/firebase notification to assignees/followers/creators
	for _, assignee := range appNotif.Assignees {
		if _, ok := duplicateMasker[assignee.UserID]; !ok {
			appNotif.Send(ctx, assignee, notificationType, db, firebaseSDKPath)
			duplicateMasker[assignee.UserID] = true
		}
	}

	for _, follower := range appNotif.Followers {
		if _, ok := duplicateMasker[follower.UserID]; !ok {
			appNotif.Send(ctx, follower, notificationType, db, firebaseSDKPath)
			duplicateMasker[follower.UserID] = true
		}
	}

	appNotifItem, err := appNotif.Save(ctx, notificationType, db)
	if err != nil {
		log.Println("***>***> addNotification: unexpected/unhandled error occurred while saving notification for conversation. error:", err)
		return err
	}

	log.Println(appNotifItem)
	// notifEntity, err := entity.Retrieve(ctx, appNotifItem.AccountID, appNotifItem.EntityID, db)
	// if err != nil {
	// 	return err
	// }
	// valueAddedFields := notifEntity.ValueAdd(appNotifItem.Fields())

	// valueAddedFields = appendTimers(appNotifItem.CreatedAt, util.ConvertMilliToTime(appNotifItem.UpdatedAt), appNotifItem.UserID, valueAddedFields)
	// err = j.actOnRedisGraph(ctx, appNotifItem.AccountID, appNotifItem.EntityID, appNotifItem.ID, nil, valueAddedFields, j.DB, j.Rpool)
	// if err != nil {
	// 	return err
	// }
	return nil
}
