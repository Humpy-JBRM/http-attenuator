package redis

import (
	"fmt"
	"log"
	"sync"
	"time"

	"http-attenuator/data"
	config "http-attenuator/facade/config"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/redigo"
	"github.com/gomodule/redigo/redis"
)

var poolQueue *redis.Pool
var redSyncQueue *redsync.Redsync
var redisOnceQueue sync.Once

func getPoolQueue() *redis.Pool {
	redisOnceQueue.Do(func() {
		InitialiseRedisQueue()
	})

	return poolQueue
}

// RedisDo wraps redis.Conn.Do() returning the reply
func QueueDo(command string, args ...interface{}) (interface{}, error) {
	conn := getPoolQueue().Get()

	if err := conn.Err(); err != nil {
		return nil, fmt.Errorf("ERROR|redisDoQueue(%s %v)|could not get connection from redis pool| %v", command, args, err)
	}

	defer conn.Close()

	return conn.Do(command, args...)
}

func NewMutexQueue(name string) *redsync.Mutex {
	return redSyncQueue.NewMutex(name)
}

func InitialiseRedisQueue() {
	server, err := config.Config().GetString(data.CONF_REDIS_HOST)
	if err != nil || server == "" {
		log.Fatalf("FATAL|InitialiseRedisQueue()|'%s' not set|%v", data.CONF_REDIS_HOST, err)
	}

	// Default values
	pass, err := config.Config().GetSecret(data.CONF_QUEUE_REDIS_PASSWORD)
	if err != nil {
		log.Fatalf("FATAL|InitialiseRedisQueue()|Could not read redis password from '%s'|%v", data.CONF_QUEUE_REDIS_PASSWORD, err)
	}

	maxIdle := 600
	maxActive := 10
	idleTimeout := 15 * time.Second
	if idle, err := config.Config().GetInt(data.CONF_QUEUE_REDIS_MAX_IDLE); err != nil && idle > 0 {
		maxIdle = int(idle)
	}
	if active, err := config.Config().GetInt(data.CONF_QUEUE_REDIS_MAX_ACTIVE); err != nil && active > 0 {
		maxActive = int(active)
	}

	dialFunc := func() (redis.Conn, error) {
		return redis.Dial("tcp", server,
			redis.DialPassword(pass),
			redis.DialDatabase(0),
		)
	}
	if pass == "" {
		dialFunc = func() (redis.Conn, error) {
			return redis.Dial("tcp", server,
				redis.DialDatabase(0),
			)
		}
	}
	poolQueue = &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: idleTimeout,
		Dial:        dialFunc,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	redSyncQueue = redsync.New(redigo.NewPool(poolQueue))

	log.Printf("made inital redis connection (%s)", server)
}
