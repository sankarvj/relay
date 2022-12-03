package database

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
)

const (
	SocketNameSpace = "SocketAuth"
	EntityNameSpace = "Entity"
	ItemIDNameSpace = "ItemID"
)

type SecDB struct {
	redisGraphPool  *redis.Pool
	redisCachePool  *redis.Pool
	redisPubSubPool *redis.Pool
}

func Init(graphPool, cachePool, pubsubPool *redis.Pool) *SecDB {
	return &SecDB{
		redisGraphPool:  graphPool,
		redisCachePool:  cachePool,
		redisPubSubPool: pubsubPool,
	}
}

func (sdb *SecDB) Scheduler() *redis.Pool {
	return sdb.redisCachePool
}

func (sdb *SecDB) Cache() *redis.Pool {
	return sdb.redisCachePool
}

func (sdb *SecDB) GraphPool() *redis.Pool {
	return sdb.redisGraphPool
}

func (sdb *SecDB) PubSubPool() *redis.Pool {
	return sdb.redisPubSubPool
}

func (sdb *SecDB) SetSocketAuthToken(key, value string) error {
	conn := sdb.redisCachePool.Get()
	defer conn.Close()

	err := conn.Send("SET", fmt.Sprintf("%s:%s", SocketNameSpace, key), value)
	if err != nil {
		return err
	}

	_, err = conn.Do("EXPIRE", fmt.Sprintf("%s:%s", SocketNameSpace, key), 5)
	return err
}

func (sdb *SecDB) RetriveSocketAuthToken(key string) (string, error) {
	conn := sdb.redisCachePool.Get()
	defer conn.Close()
	return redis.String(conn.Do("GET", fmt.Sprintf("%s:%s", SocketNameSpace, key)))
}

func (sdb *SecDB) SetUserToken(key string, value []byte) error {
	conn := sdb.redisCachePool.Get()
	defer conn.Close()
	err := conn.Send("SET", key, value)
	return err
}

func (sdb *SecDB) GetUserToken(key string) (string, error) {
	conn := sdb.redisCachePool.Get()
	defer conn.Close()
	return redis.String(conn.Do("GET", key))

}

func (sdb *SecDB) SetEntity(key string, encodedEntity []byte) error {
	conn := sdb.redisCachePool.Get()
	defer conn.Close()

	err := conn.Send("SET", fmt.Sprintf("%s:%s", EntityNameSpace, key), encodedEntity)
	if err != nil {
		return err
	}

	_, err = conn.Do("EXPIRE", fmt.Sprintf("%s:%s", EntityNameSpace, key), 25)
	return err
}

func (sdb *SecDB) RetriveEntity(key string) (string, error) {
	conn := sdb.redisCachePool.Get()
	defer conn.Close()
	return redis.String(conn.Do("GET", fmt.Sprintf("%s:%s", EntityNameSpace, key)))
}

func (sdb *SecDB) ResetEntity(key string) (string, error) {
	conn := sdb.redisCachePool.Get()
	defer conn.Close()
	return redis.String(conn.Do("DELETE", fmt.Sprintf("%s:%s", EntityNameSpace, key)))
}

func (sdb *SecDB) SetItemID(key string, itemID string) error {
	conn := sdb.redisCachePool.Get()
	defer conn.Close()

	err := conn.Send("SET", fmt.Sprintf("%s:%s", ItemIDNameSpace, key), itemID)
	if err != nil {
		return err
	}

	_, err = conn.Do("EXPIRE", fmt.Sprintf("%s:%s", ItemIDNameSpace, key), 25)
	return err
}

func (sdb *SecDB) RetriveItemID(key string) (string, error) {
	conn := sdb.redisCachePool.Get()
	defer conn.Close()
	return redis.String(conn.Do("GET", fmt.Sprintf("%s:%s", ItemIDNameSpace, key)))
}
