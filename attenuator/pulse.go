package attenuator

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// A Pulse emits a heartbeat every N milliseconds.
//
// How this value is obtained from config is an exercise for
// the interested reader.
//
// For this exerecise, our rate limit is a maximum of 2 per second,
// so the light goes green every 500 ms.
type Pulse interface {
	// wait for the next heartbeat
	WaitForNext()

	// pause heartbeats for a specified amount of time
	SetPauseForDuration(duration time.Duration)

	// pause hartbeats until a particular wallclock time is reached
	SetPauseUntil(wallclock time.Time)
}

type PulseImpl struct {
	name       string
	numWorkers int
	maxHertz   float64
	pulseChan  chan (bool)

	// TODO(john): deal with current flow rate
	currentRateHertz float64

	// TODO(john): deal with maintaining target flow rate
	// This might be very useful for load testing
	targetRateHertz float64

	waitUntil *time.Time
}

var pulses = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "pulse",
		Help:      "The pulse heartbeats, keyed by pulse name",
	},
	[]string{"name", "hertz"},
)

var pulseRegistry map[string]Pulse = map[string]Pulse{}
var prMutex sync.RWMutex

func GetPulse(name string) Pulse {
	prMutex.RLock()
	defer prMutex.RUnlock()
	return pulseRegistry[strings.ToLower(name)]
}

func NewPulse(name string, numWorkers int, maxHertz float64, targetHertz float64) (Pulse, error) {
	prMutex.RLock()
	if _, exists := pulseRegistry[strings.ToLower(name)]; exists {
		prMutex.RUnlock()
		return nil, fmt.Errorf("Pulse '%s' already exists", name)
	}
	prMutex.RUnlock()

	pulse := &PulseImpl{
		name:            name,
		numWorkers:      numWorkers,
		maxHertz:        maxHertz,
		targetRateHertz: targetHertz,
		pulseChan:       make(chan bool, numWorkers),
	}
	if targetHertz <= 0 {
		pulse.targetRateHertz = pulse.maxHertz
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
				pulses.WithLabelValues(p.name, fmt.Sprintf("%.2f", p.maxHertz)).Inc()
				p.pulseChan <- true
				continue
			}

			// If we are to wait until a specified time, then do so
			if p.waitUntil != nil {
				sleepDurationNano := p.waitUntil.UnixNano() - time.Now().UnixNano()
				if sleepDurationNano > 0 {
					time.Sleep(time.Duration(sleepDurationNano) * time.Nanosecond)
				}
				pulses.WithLabelValues(p.name, fmt.Sprintf("%.2f", p.maxHertz)).Inc()
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

func (p *PulseImpl) WaitForNext() {
	// log.Println("Waiting for traffic light")
	<-p.pulseChan
	// log.Println("Got traffic light")
}

func (p *PulseImpl) SetPauseForDuration(duration time.Duration) {
	t := time.Now().Add(duration)
	p.waitUntil = &t
}

func (p *PulseImpl) SetPauseUntil(wallclock time.Time) {
	p.waitUntil = &wallclock
}
