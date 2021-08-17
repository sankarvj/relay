package redisdb

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
)

const (
	SocketNameSpace = "SocketAuth"
)

func RedisSet(rp *redis.Pool, key, value string) error {
	conn := rp.Get()
	defer conn.Close()

	err := conn.Send("SET", fmt.Sprintf("%s:%s", SocketNameSpace, key), value)
	if err != nil {
		return err
	}

	_, err = conn.Do("EXPIRE", fmt.Sprintf("%s:%s", SocketNameSpace, key), 5)
	return err
}

func RedisGet(rp *redis.Pool, key string) (string, error) {
	conn := rp.Get()
	defer conn.Close()

	return redis.String(conn.Do("GET", fmt.Sprintf("%s:%s", SocketNameSpace, key)))

}
