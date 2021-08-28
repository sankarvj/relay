package conversation

import (
	"github.com/gomodule/redigo/redis"
)

type Hub struct {
	Clients    map[string]map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
}

// NewInstanceHub creates a new hub for an aws-instance
func NewInstanceHub() *Hub {
	return &Hub{
		Clients:    make(map[string]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Run our websocket server, accepting various requests
func (hub *Hub) Run(rp *redis.Pool) {
	go hub.listenPubSubChannel(rp)
	for {
		select {
		case client := <-hub.Register:
			hub.registerClient(client, rp)

		case client := <-hub.Unregister:
			hub.unregisterClient(client, rp)
		}
	}
}

func (hub *Hub) registerClient(client *Client, rp *redis.Pool) error {
	if roomClients, ok := hub.Clients[client.room]; ok {
		roomClients[client] = true
	} else {
		roomClients = make(map[*Client]bool, 0)
		roomClients[client] = true
		hub.Clients[client.room] = roomClients
	}
	message := NewMessage(UserJoinedAction, ViewModelConversation{}, client.room, client.user, client.id)
	err := hub.publishClientJoined(message, rp)
	return err
}

func (hub *Hub) unregisterClient(client *Client, rp *redis.Pool) {
	if roomClients, ok := hub.Clients[client.room]; ok {
		delete(roomClients, client)
		hub.Clients[client.room] = roomClients
		message := NewMessage(UserLeftAction, ViewModelConversation{}, client.room, client.user, client.id)
		hub.publishClientLeft(message, rp)
	}
}

func (hub *Hub) handleUserJoined(msg Message) {
	//DEBUGGING LOG fmt.Printf("message handleUserJoined ---------------------------------------> %+v", msg)
}

func (hub *Hub) handleUserLeft(msg Message) {
	//DEBUGGING LOG fmt.Printf("message handleUserLeft ---------------------------------------> %+v", msg)
}

func (hub *Hub) handleIncomingMessage(msg Message) {
	//DEBUGGING LOG fmt.Printf("message handleIncomingMessage ---------------------------------------> %+v", msg)
	if roomClients, ok := hub.Clients[msg.Room]; ok {
		for client := range roomClients {
			//if client.id != msg.ClientID {
			client.send <- msg.Payload.encode()
			//}
		}
	}

}
