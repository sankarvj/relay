package secondarydb

import "github.com/gomodule/redigo/redis"

type SecDB struct {
	RedisGraphPool  *redis.Pool
	RedisCachePool  *redis.Pool
	RedisPubSubPool *redis.Pool
}
