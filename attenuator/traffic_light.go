package attenuator

import (
	"log"
	"strings"
)

type TrafficLight interface {
	WaitForGreen(attemptNumber int)
}

type TrafficLightImpl struct {
	Name  string
	pulse Pulse
}

func (t *TrafficLightImpl) WaitForGreen(attemptNumber int) {
	// return straigh away of no pulse has been configured
	if t.pulse == nil {
		return
	}

	t.pulse.WaitForNext()
}

// traffic light singleton
//
// THIS is what makes it available globally.
var trafficLightInstance map[string]TrafficLight = map[string]TrafficLight{}

func init() {
	RegisterTrafficLight(&TrafficLightImpl{
		Name:  "",
		pulse: nil,
	})

	RegisterTrafficLight(&TrafficLightImpl{
		Name: "whisper",
		// assume 1 worker, pulse is 2 per second
		//
		// TODO(james): make the first parameter equal to the number of whisper transcribe workers
		//
		// TODO(james): make the second parameter equal to the GLOBAL RATE LIMIT.
		//              NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE
		//
		//				THIS VALUE IS IN HERTZ
		//
		//				A VALUE OF ZERO DISABLES THE RATE LIMIT
		pulse: NewPulse(10, 1, 0),
	})
}

func RegisterTrafficLight(t *TrafficLightImpl) {
	trafficLightInstance[strings.ToLower(t.Name)] = t
}

func WaitForGreen(trafficLightName string, attemptNumber int) {
	log.Printf("Waiting for %s (attempt %d)", trafficLightName, attemptNumber)
	if tl := trafficLightInstance[strings.ToLower(trafficLightName)]; tl != nil {
		tl.WaitForGreen(attemptNumber)
	} else {
		panic("*")
	}
}
