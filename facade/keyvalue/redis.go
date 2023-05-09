package facade

import (
	"fmt"
	"http-attenuator/data"
	config "http-attenuator/facade/config"
	"http-attenuator/util"
	"reflect"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

type RedisKeyValue struct {
	redisHost string
	redisPool *redis.Pool
}

func NewRedisKeyValue() (KeyValue, error) {
	// Get the redis config values
	redisHost, err := config.Config().GetString(data.CONF_REDIS_HOST)
	if err != nil {
		return nil, fmt.Errorf("NewRedisKeyValue(): %s", err.Error())
	}
	poolSize, err := config.Config().GetInt(data.CONF_REDIS_POOLSIZE)
	if err != nil {
		return nil, fmt.Errorf("NewRedisKeyValue(): %s", err.Error())
	}
	timeout, err := config.Config().GetInt(data.CONF_REDIS_TIMEOUT)
	if err != nil {
		return nil, fmt.Errorf("NewRedisKeyValue(): %s", err.Error())
	}
	return &RedisKeyValue{
		redisHost: redisHost,
		redisPool: util.NewRedisConnectionPool(redisHost, int(poolSize), time.Duration(timeout)*time.Millisecond),
	}, nil
}

func (kv *RedisKeyValue) Set(key string, value any) error {
	conn := kv.redisPool.Get()
	if err := conn.Err(); err != nil {
		return fmt.Errorf("redis.Set(%s): %s", key, err)
	}
	defer conn.Close()
	_, err := conn.Do(
		"SET",
		key,
		fmt.Sprint(value),
	)
	if err != nil {
		return fmt.Errorf("redis.Set(%s): %s", key, err)
	}
	return nil
}

func (kv *RedisKeyValue) GetString(key string) (string, error) {
	conn := kv.redisPool.Get()
	if err := conn.Err(); err != nil {
		return "", fmt.Errorf("redis.GetString(%s): %s", key, err)
	}
	defer conn.Close()
	v, err := conn.Do(
		"GET",
		key,
	)
	if err != nil {
		return "", fmt.Errorf("redis.GetString(%s): %s", key, err)
	}
	return string(v.([]byte)), nil
}

func (kv *RedisKeyValue) GetInt(key string) (int64, error) {
	conn := kv.redisPool.Get()
	if err := conn.Err(); err != nil {
		return 0, fmt.Errorf("redis.GetInt(%s): %s", key, err)
	}
	defer conn.Close()
	v, err := conn.Do(
		"GET",
		key,
	)
	if err != nil {
		return 0, fmt.Errorf("redis.GetInt(%s): %s", key, err)
	}
	if v == nil {
		return 0, nil
	}

	numericValue, err := strconv.ParseInt(string(v.([]byte)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("redis.GetInt(%s): cannot convert %s to int64: %s", key, reflect.TypeOf(v), err.Error())
	}
	return numericValue, nil
}

func (kv *RedisKeyValue) GetFloat(key string) (float64, error) {
	conn := kv.redisPool.Get()
	if err := conn.Err(); err != nil {
		return 0, fmt.Errorf("redis.GetFloat(%s): %s", key, err)
	}
	defer conn.Close()
	v, err := conn.Do(
		"GET",
		key,
	)
	if err != nil {
		return 0, fmt.Errorf("redis.GetFloat(%s): %s", key, err)
	}
	if v == nil {
		return 0, nil
	}

	numericValue, err := strconv.ParseFloat(string(v.([]byte)), 64)
	if err != nil {
		return 0, fmt.Errorf("redis.GetFloat(%s): cannot convert %s to float64: %s", key, reflect.TypeOf(v), err.Error())
	}
	return numericValue, nil
}

func (kv *RedisKeyValue) GetBool(key string) (bool, error) {
	conn := kv.redisPool.Get()
	if err := conn.Err(); err != nil {
		return false, fmt.Errorf("redis.GetBool(%s): %s", key, err)
	}
	defer conn.Close()
	v, err := conn.Do(
		"GET",
		key,
	)
	if err != nil {
		return false, fmt.Errorf("redis.GetBool(%s): %s", key, err)
	}
	if v == nil {
		return false, nil
	}

	return string(v.([]byte)) == "true", nil
}

func (kv *RedisKeyValue) Delete(key string) error {
	conn := kv.redisPool.Get()
	if err := conn.Err(); err != nil {
		return fmt.Errorf("redis.Delete(%s): %s", key, err)
	}
	defer conn.Close()
	_, err := conn.Do(
		"DEL",
		key,
	)
	if err != nil {
		return fmt.Errorf("redis.Delete(%s): %s", key, err)
	}
	return nil
}

func (kv *RedisKeyValue) Add(key string, delta int64) (int64, error) {
	conn := kv.redisPool.Get()
	if err := conn.Err(); err != nil {
		return 0, fmt.Errorf("redis.Add(%s): %s", key, err)
	}
	defer conn.Close()
	v, err := conn.Do(
		"INCRBY",
		key,
		delta,
	)
	if err != nil {
		return 0, fmt.Errorf("redis.Add(%s): %s", key, err)
	}
	if v == nil {
		return 0, nil
	}

	numericValue, err := strconv.ParseInt(fmt.Sprint(v), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("redis.Add(%s): cannot increase by %d: %s", key, delta, err.Error())
	}
	return numericValue, nil
}

func (kv *RedisKeyValue) Dec(key string, delta int64) (int64, error) {
	conn := kv.redisPool.Get()
	if err := conn.Err(); err != nil {
		return 0, fmt.Errorf("redis.Dec(%s): %s", key, err)
	}
	defer conn.Close()
	v, err := conn.Do(
		"DECRBY",
		key,
		delta,
	)
	if err != nil {
		return 0, fmt.Errorf("redis.Dec(%s): %s", key, err)
	}
	if v == nil {
		return 0, nil
	}

	numericValue, err := strconv.ParseInt(fmt.Sprint(v), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("redis.Dec(%s): cannot decrease by %d: %s", key, delta, err.Error())
	}
	return numericValue, nil
}
