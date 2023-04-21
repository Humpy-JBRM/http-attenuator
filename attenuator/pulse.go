package attenuator

import (
	"time"
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

func NewPulse(numWorkers int, maxHertz float64, targetHertz float64) Pulse {
	pulse := &PulseImpl{
		numWorkers:      numWorkers,
		maxHertz:        maxHertz,
		targetRateHertz: targetHertz,
		pulseChan:       make(chan bool, numWorkers),
	}
	if targetHertz <= 0 {
		pulse.targetRateHertz = pulse.maxHertz
	}

	// Kick off the pulse
	go func(p *PulseImpl) {
		if maxHertz <= 0 {
			return
		}

		sleepTimeMillis := 1000 / maxHertz
		for {
			if sleepTimeMillis <= 0 {
				// always a green light
				p.pulseChan <- true
				continue
			}

			// If we are to wait until a specified time, then do so
			if p.waitUntil != nil {
				sleepDurationNano := p.waitUntil.UnixNano() - time.Now().UnixNano()
				if sleepDurationNano > 0 {
					time.Sleep(time.Duration(sleepDurationNano) * time.Nanosecond)
				}
				p.pulseChan <- true
				p.waitUntil = nil
				continue
			}

			// wait for the heartbeat
			time.Sleep(time.Duration(sleepTimeMillis) * time.Millisecond)
			p.pulseChan <- true
		}
	}(pulse)
	return pulse
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
