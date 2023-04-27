package util

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

func NewRedisConnectionPool(hostPort string, size int, timeout time.Duration) *redis.Pool {
	dialFunc := func() (redis.Conn, error) {
		return redis.Dial("tcp", hostPort,
			redis.DialDatabase(0),
		)
	}
	poolQueue := &redis.Pool{
		MaxIdle:     size,
		MaxActive:   size,
		IdleTimeout: timeout,
		Dial:        dialFunc,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	return poolQueue
}
