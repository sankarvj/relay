package conversation

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

const (
	// Max wait time when writing message to peer
	writeWait = 10 * time.Second

	// Max time till next pong from peer
	pongWait = 60 * time.Second

	// Send ping interval, must be less then pong wait time
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 10000
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client represents the websocket client at the server
type Client struct {
	// The actual websocket connection.
	conn   *websocket.Conn
	hub    *Hub
	id     string
	base   string
	room   string
	user   string
	name   string
	avatar string
	send   chan []byte
}

func NewClient(conn *websocket.Conn, hub *Hub, id, base, room, user, email, name, avatar string) *Client {

	if avatar == "" {
		avatar = util.Avatar(email)
	}

	return &Client{
		conn:   conn,
		hub:    hub,
		id:     id,
		base:   base,
		room:   room,
		user:   user,
		name:   name,
		avatar: avatar,
		send:   make(chan []byte, 256),
	}
}

func (client *Client) disconnect() {
	client.hub.Unregister <- client
	close(client.send)
	client.conn.Close()
}

func (client *Client) ReadPump(rp *redis.Pool, messageChan chan Message) {
	defer func() {
		client.disconnect()
	}()

	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error { client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// Start endless read loop, waiting for messages from client
	for {
		_, jsonMessage, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("***> unexpected/unhandled error occurred when closing the read connection. error:", err)
			}
			break
		}

		var viewModelConv ViewModelConversation
		if err := json.Unmarshal(jsonMessage, &viewModelConv); err != nil {
			log.Println("***> unexpected/unhandled error occurred when unmarshal message. error:", err)
		}

		viewModelConv.ID = uuid.New().String()
		viewModelConv.UserID = client.user
		viewModelConv.UserAvatar = client.avatar
		viewModelConv.UserName = client.name

		//sending the message to the handler for storing it in the DB
		message := NewMessage(SendMessageAction, viewModelConv, client.base, client.room, client.user, client.id)
		messageChan <- *message

		//sending the message to the pub/sub
		err = client.hub.publishReceivedMessage(message, rp)
		if err != nil {
			log.Println("***> unexpected/unhandled error occurred when publishing the message to redis. error:", err)
		}
	}
}

func (client *Client) WritePump(rp *redis.Pool) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()
	for {
		select {
		case messageBytes, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The WsServer closed the channel.
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(messageBytes)

			// Attach queued chat messages to the current websocket message.
			n := len(client.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-client.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
