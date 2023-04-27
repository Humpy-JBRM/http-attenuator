package facade

import (
	"fmt"
	"http-attenuator/data"
	config "http-attenuator/facade/config"
	"http-attenuator/util"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

type RedisPulseImpl struct {
	PulseImpl
	redisHost string
	redisPool *redis.Pool
}

func NewRedisPulse(name string, numWorkers int, maxHertz float64, targetHertz float64, redisHost string) (Pulse, error) {
	// Get the redis config values
	redisHost, err := config.Config().GetString(data.CONF_REDIS_HOST)
	if err != nil {
		return nil, fmt.Errorf("NewRedisPulse(%s): %s", name, err.Error())
	}
	poolSize, err := config.Config().GetInt(data.CONF_REDIS_POOLSIZE)
	if err != nil {
		return nil, fmt.Errorf("NewRedisPulse(%s): %s", name, err.Error())
	}
	timeout, err := config.Config().GetInt(data.CONF_REDIS_TIMEOUT)
	if err != nil {
		return nil, fmt.Errorf("NewRedisPulse(%s): %s", name, err.Error())
	}

	prMutex.RLock()
	if _, exists := pulseRegistry[strings.ToLower(name)]; exists {
		prMutex.RUnlock()
		return nil, fmt.Errorf("Pulse '%s' already exists", name)
	}
	prMutex.RUnlock()

	pulse := &RedisPulseImpl{
		PulseImpl: PulseImpl{
			name:            name,
			maxInflight:     numWorkers,
			maxHertz:        maxHertz,
			targetRateHertz: targetHertz,
		},
		redisHost: redisHost,
		redisPool: util.NewRedisConnectionPool(redisHost, int(poolSize), time.Duration(timeout)*time.Millisecond),
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
	conn := p.redisPool.Get()
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
	conn := p.redisPool.Get()
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
