package facade

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

type RedisPulseImpl struct {
	PulseImpl
	redisHost string
}

var poolQueue *redis.Pool
var redisOnceQueue sync.Once

func getConnectionPool() *redis.Pool {
	redisOnceQueue.Do(func() {
		// TODO(john): get these from config
		maxIdle := 10
		maxActive := 10
		idleTimeout := 10 * time.Second

		dialFunc := func() (redis.Conn, error) {
			return redis.Dial("tcp", "localhost:6379",
				redis.DialDatabase(0),
			)
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
	})

	return poolQueue
}

func NewRedisPulse(name string, numWorkers int, maxHertz float64, targetHertz float64, redisHost string) (Pulse, error) {
	prMutex.RLock()
	if _, exists := pulseRegistry[strings.ToLower(name)]; exists {
		prMutex.RUnlock()
		return nil, fmt.Errorf("Pulse '%s' already exists", name)
	}
	prMutex.RUnlock()

	pulse := &RedisPulseImpl{
		PulseImpl: PulseImpl{
			name:            name,
			numWorkers:      numWorkers,
			maxHertz:        maxHertz,
			targetRateHertz: targetHertz,
		},
		redisHost: redisHost,
	}
	if targetHertz < 0 {
		pulse.targetRateHertz = 0
	}
	prMutex.Lock()
	pulseRegistry[strings.ToLower(name)] = pulse
	prMutex.Unlock()

	// Kick off the pulse
	//
	// All we do is keep adding to a redis queue
	go func(p *RedisPulseImpl) {
		if maxHertz <= 0 {
			return
		}

		sleepTimeMillis := 1000 / maxHertz
		for {
			if sleepTimeMillis <= 0 {
				// always a green light
				pulses.WithLabelValues(p.name, "redis", fmt.Sprintf("%.2f", p.maxHertz)).Inc()
				p.sendPulse()
				continue
			}

			// If we are to wait until a specified time, then do so
			if p.waitUntil != nil {
				sleepDurationNano := p.waitUntil.UnixNano() - time.Now().UTC().UnixNano()
				if sleepDurationNano > 0 {
					time.Sleep(time.Duration(sleepDurationNano) * time.Nanosecond)
				}
				pulses.WithLabelValues(p.name, "redis", fmt.Sprintf("%.2f", p.maxHertz)).Inc()
				p.sendPulse()
				p.waitUntil = nil
				continue
			}

			// wait for the heartbeat
			time.Sleep(time.Duration(sleepTimeMillis) * time.Millisecond)
			p.sendPulse()
		}
	}(pulse)
	return pulse, nil
}

func (p *RedisPulseImpl) WaitForNext() error {
	// log.Println("Waiting for traffic light")
	// log.Println("Got traffic light")
	conn := getConnectionPool().Get()
	if err := conn.Err(); err != nil {
		return fmt.Errorf("redis.WaitForNext(): %s", err)
	}
	defer conn.Close()

	pulseWaitTime.WithLabelValues(p.name, "naive").Inc()
	_, err := conn.Do(
		"BRPOP",
		p.name,
	)
	if err != nil {
		return fmt.Errorf("redis.WaitForNext(): %s", err)
	}
	pulseDrain.WithLabelValues(p.name, "naive").Inc()
	return nil
}

// There is an infintessimally small chance that the
// queue gets read between us checking the size and writing
// to the queue.
//
// This is acceptable, the law of large numbers will smooth
// it out
func (p *RedisPulseImpl) sendPulse() error {
	conn := getConnectionPool().Get()
	if err := conn.Err(); err != nil {
		return fmt.Errorf("redis.sendPulse(): %s", err)
	}
	defer conn.Close()
	lenValue, err := conn.Do(
		"LLEN",
		p.name,
	)
	if err != nil {
		return fmt.Errorf("redis.sendPulse(): %s", err)
	}
	len, err := strconv.ParseInt(fmt.Sprint(lenValue), 10, 64)
	if err != nil {
		return fmt.Errorf("redis.sendPulse(): %s", err)
	}
	if len == 0 {
		_, err = conn.Do(
			"LPUSH",
			p.name,
			fmt.Sprint(time.Now().UTC().UnixMilli()),
		)
		if err != nil {
			return fmt.Errorf("redis.sendPulse(): %s", err)
		}
		pulseSink.WithLabelValues(p.name, "naive").Inc()
	}

	return nil
}
