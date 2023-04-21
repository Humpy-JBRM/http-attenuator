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

	pulse, err := NewPulse(
		"www.google.com",
		10, // number of workers / size of pulses channel
		1,  // pulse hertz
		0,  // target hertz (this is TODO)
	)
	if err != nil {
		log.Fatal(err)
	}
	RegisterTrafficLight(&TrafficLightImpl{
		Name: "www.google.com",
		// assume 10 workers, pulse is 1 per second
		pulse: pulse,
	})
}

func RegisterTrafficLight(t *TrafficLightImpl) {
	trafficLightInstance[strings.ToLower(t.Name)] = t
}

func WaitForGreen(trafficLightName string, attemptNumber int) {
	if tl := trafficLightInstance[strings.ToLower(trafficLightName)]; tl != nil {
		log.Printf("Waiting for %s (attempt %d)", trafficLightName, attemptNumber)
		tl.WaitForGreen(attemptNumber)
		return
	}

	// No traffic light / no attenuation
	log.Printf("No traffic light for %s (attempt %d)", trafficLightName, attemptNumber)
}
