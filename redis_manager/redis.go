package redis_manager

import (
	"github.com/garyburd/redigo/redis"
	"time"
)

var Pool *redis.Pool

func InitRedis(addr string, idleConn, maxConn int, idleTimeout time.Duration) {
	Pool = &redis.Pool{
		MaxIdle:     idleConn,
		MaxActive:   maxConn,
		IdleTimeout: idleTimeout,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr)
		},
	}
	return
}

func Get(key string) (value string, err error) {
	conn := Pool.Get()
	defer conn.Close()

	value, err = redis.String(conn.Do("Get", key))
	return
}

