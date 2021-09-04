package conversation

import (
	"encoding/json"

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

func (hub *Hub) publishClientJoined(message *Message, rp *redis.Pool) error {
	conn := rp.Get()
	defer conn.Close()

	_, err := conn.Do("PUBLISH", PubSubGeneralChannel, message.encode())
	return err
}

func (hub *Hub) publishClientLeft(message *Message, rp *redis.Pool) error {
	conn := rp.Get()
	defer conn.Close()

	_, err := conn.Do("PUBLISH", PubSubGeneralChannel, message.encode())
	return err
}

func (hub *Hub) publishReceivedMessage(message *Message, rp *redis.Pool) error {
	conn := rp.Get()
	defer conn.Close()
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
			//DEBUGGING LOG fmt.Printf("redis.Message  -----------------> channel: %s: message: %s\n", v.Channel, v.Data)
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
			//DEBUGGING LOG fmt.Printf("redis.Subscription -----------------> %s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			return v
		}
	}

}
