package facade

import "time"

// A Pulse emits a heartbeat every N milliseconds.
//
// How this value is obtained from config is an exercise for
// the interested reader.
type Pulse interface {
	// wait for the next heartbeat
	WaitForNext() error

	// pause heartbeats for a specified amount of time
	SetPauseForDuration(duration time.Duration)

	// pause hartbeats until a particular wallclock time is reached
	SetPauseUntil(wallclock time.Time)
}

type TrafficLight interface {
	WaitForGreen()
}
