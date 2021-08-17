package event

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/websocket"
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
	conn *websocket.Conn
	hub  *Hub
	id   string
	room string
	user string
	send chan []byte
}

func NewClient(conn *websocket.Conn, hub *Hub, id, room, user string) *Client {
	return &Client{
		conn: conn,
		hub:  hub,
		id:   id,
		room: room,
		user: user,
		send: make(chan []byte, 256),
	}
}

func (client *Client) disconnect() {
	client.hub.Unregister <- client
	close(client.send)
	client.conn.Close()
}

func (client *Client) ReadPump(rp *redis.Pool) {
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
				log.Println("unexpected error occurred when closing the read connection. error:", err)
			}
			break
		}

		var vmMessage ViewModelMessage
		if err := json.Unmarshal(jsonMessage, &vmMessage); err != nil {
			log.Println("unexpected error occurred when unmarshal message. error:", err)
		}

		err = client.hub.publishReceivedMessage(client.id, client.user, client.room, vmMessage, rp)
		if err != nil {
			log.Println("unexpected error occurred when publishing the message to redis. error:", err)
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
