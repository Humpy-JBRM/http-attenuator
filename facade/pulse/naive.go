package facade

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PulseImpl struct {
	name        string
	maxInflight int
	maxHertz    float64
	pulseChan   chan (bool)

	// TODO(john): deal with current flow rate
	currentRateHertz float64

	// TODO(john): deal with maintaining target flow rate
	// This might be very useful for load testing
	targetRateHertz float64

	waitUntil *time.Time

	// requests currently in flight
	inflight chan bool
}

var pulses = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "pulse",
		Help:      "The pulse heartbeats, keyed by pulse name and type",
	},
	[]string{"name", "type", "hertz"},
)
var pulseWaitTime = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "pulse_wait_time",
		Help:      "Time spent waiting for pulses",
	},
	[]string{"name", "type"},
)
var pulseSink = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "pulse_sink",
		Help:      "Number of pulses added",
	},
	[]string{"name", "type"},
)
var pulseDrain = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "pulse_drain",
		Help:      "Pulses fetched",
	},
	[]string{"name", "type"},
)

var pulseRegistry map[string]Pulse = map[string]Pulse{}
var prMutex sync.RWMutex

func GetPulse(name string) Pulse {
	prMutex.RLock()
	defer prMutex.RUnlock()
	return pulseRegistry[strings.ToLower(name)]
}

func NewPulse(name string, maxInflight int, maxHertz float64) (Pulse, error) {
	prMutex.RLock()
	if _, exists := pulseRegistry[strings.ToLower(name)]; exists {
		prMutex.RUnlock()
		return nil, fmt.Errorf("Pulse '%s' already exists", name)
	}
	prMutex.RUnlock()

	pulse := &PulseImpl{
		name:        name,
		maxInflight: maxInflight,
		maxHertz:    maxHertz,
		pulseChan:   make(chan bool, 1),
		inflight:    make(chan bool, maxInflight),
	}
	prMutex.Lock()
	pulseRegistry[strings.ToLower(name)] = pulse
	prMutex.Unlock()

	// Kick off the pulse
	go func(p *PulseImpl) {
		if maxHertz <= 0 {
			return
		}

		sleepTimeMillis := 1000 / maxHertz
		for {
			if sleepTimeMillis <= 0 {
				// always a green light
				pulses.WithLabelValues(p.name, "naive", fmt.Sprintf("%.2f", p.maxHertz)).Inc()
				p.pulseChan <- true
				continue
			}

			// If we are to wait until a specified time, then do so
			if p.waitUntil != nil {
				sleepDurationNano := p.waitUntil.UnixNano() - time.Now().UTC().UnixNano()
				if sleepDurationNano > 0 {
					time.Sleep(time.Duration(sleepDurationNano) * time.Nanosecond)
				}
				pulses.WithLabelValues(p.name, "naive", fmt.Sprintf("%.2f", p.maxHertz)).Inc()
				p.pulseChan <- true
				p.waitUntil = nil
				continue
			}

			// wait for the heartbeat
			time.Sleep(time.Duration(sleepTimeMillis) * time.Millisecond)
			p.pulseChan <- true
		}
	}(pulse)
	return pulse, nil
}

// startInflight waits until there are < maxInflight requests
// currently in flight
func (p *PulseImpl) startInflight() error {
	// this will block until an inflight slot is available
	p.inflight <- true
	return nil
}

// finishInflight notifies the pulse that an inflight
// request has finished
func (p *PulseImpl) finishInflight() error {
	select {
	case <-p.inflight:
	default:
	}
	return nil
}

func (p *PulseImpl) WaitForNext() error {
	// Wait until the number currently in flight is lower than
	// the max allowed
	defer p.finishInflight()
	p.startInflight()
	<-p.pulseChan
	return nil
}

func (p *PulseImpl) SetPauseForDuration(duration time.Duration) {
	t := time.Now().UTC().Add(duration)
	p.waitUntil = &t
}

func (p *PulseImpl) SetPauseUntil(wallclock time.Time) {
	p.waitUntil = &wallclock
}
