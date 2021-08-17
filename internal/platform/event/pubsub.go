package event

import (
	"encoding/json"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
)

const PubSubGeneralChannel = "general"

const (
	SendMessageAction = "send-message"
	UserJoinedAction  = "user-join"
	UserLeftAction    = "user-left"
)

type Publisher struct {
	Topic string
}

func (hub *Hub) publishClientJoined(clientID, user, room string, rp *redis.Pool) error {
	conn := rp.Get()
	defer conn.Close()
	message := NewMessage(UserJoinedAction, "", room, user, clientID)
	_, err := conn.Do("PUBLISH", PubSubGeneralChannel, message.encode())
	return err
}

func (hub *Hub) publishClientLeft(clientID, user, room string, rp *redis.Pool) error {
	conn := rp.Get()
	defer conn.Close()
	message := NewMessage(UserLeftAction, "", room, user, clientID)
	_, err := conn.Do("PUBLISH", PubSubGeneralChannel, message.encode())
	return err
}

func (hub *Hub) publishReceivedMessage(clientID, user, room string, vmMessage ViewModelMessage, rp *redis.Pool) error {
	conn := rp.Get()
	defer conn.Close()
	message := NewMessage(SendMessageAction, vmMessage.Payload, room, user, clientID)
	_, err := conn.Do("PUBLISH", PubSubGeneralChannel, message.encode())
	return err
}

func (hub *Hub) listenPubSubChannel(rp *redis.Pool) error {
	conn := rp.Get()
	defer conn.Close()
	psc := redis.PubSubConn{Conn: conn}
	psc.Subscribe(PubSubGeneralChannel)

	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			fmt.Printf("redis.Message  -----------------> channel: %s: message: %s\n", v.Channel, v.Data)
			var message Message
			if err := json.Unmarshal([]byte(v.Data), &message); err != nil {
				return errors.Wrap(err, "Error on unmarshal JSON message")
			}
			switch message.Action {
			case SendMessageAction:
				hub.handleIncomingMessage(message)
			case UserJoinedAction:
				hub.handleUserJoined(message)
			case UserLeftAction:
				hub.handleUserLeft(message)
			}
		case redis.Subscription:
			fmt.Printf("redis.Subscription -----------------> %s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			return v
		}
	}

}
