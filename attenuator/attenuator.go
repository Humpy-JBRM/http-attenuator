package attenuator

import (
	"http-attenuator/circuitbreaker"
	"http-attenuator/data"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type attenuator struct {
	name         string
	maxHertz     float64
	targetHertz  float64
	workers      int
	requestQueue chan (*data.HttpRequest)
	wg           sync.WaitGroup
	workerCount  int64
	stopped      bool
	trafficLight string
}

func NewAttenuator(name string, maxHertz float64, targetHertz float64, workers int) Attenuator {
	RegisterTrafficLight(
		&TrafficLightImpl{
			Name:  name,
			pulse: NewPulse(workers, maxHertz, targetHertz),
		})

	a := &attenuator{
		name:     name,
		maxHertz: maxHertz,

		// targetHertz is not implement yet
		// it is used for auto-scaling up and down
		targetHertz:  0,
		workers:      workers,
		requestQueue: make(chan (*data.HttpRequest), 100),
		trafficLight: name,
	}

	// Kick off the attenuator workers
	for i := workers; i > 0; i-- {
		a.wg.Add(1)
		go a.worker()
	}
	return a
}

func (a *attenuator) Stop() {
	a.stopped = true
}

func (a *attenuator) next() *data.HttpRequest {
	// only happens in async mode
	// TODO(john): fix async mode
	return nil
	// log.Println("Waiting for green light")
	WaitForGreen(a.trafficLight, 1)
	// log.Println("Got green light")
	var nextRequest *data.HttpRequest
	select {
	case nextRequest = <-a.requestQueue:
		if nextRequest == nil {
			// nil request means 'no more requests'
			nextRequest = nil
		}
		// Deal with the request
		log.Printf("Processing %+v", nextRequest)

	case <-time.After(100 * time.Millisecond):
		if a.stopped {
			// Attenuator has been stopped
			nextRequest = nil
		}
	}

	return nextRequest
}

func (a *attenuator) worker() {
	thisWorker := a.workerCount
	log.Printf("Attenuator: starting worker %d", thisWorker)
	defer func() {
		a.wg.Done()
		log.Printf("Attenuator: terminating worker %d", thisWorker)

		// one less worker.
		// This allows the attenuator to spin up more workers
		// if we'e not hitting the target rate
		//
		// TODO(john): reach target rate
		atomic.AddInt64(&(a.workerCount), -1)
	}()

	for {
		next := a.next()
		if next == nil && a.stopped {
			return
		}
	}
}

func (a *attenuator) DoSync(req *data.HttpRequest) (*data.HttpResponse, error) {
	if req.Client == nil {
		req.Client = &http.Client{}
	}

	// wait for green light
	WaitForGreen(a.name, 1)

	// Do the request
	switch strings.ToLower(req.Req.Method) {
	case "get":
		code, body, headers, err := circuitbreaker.NewCircuitBreaker().HttpGet(req.Req.URL.String(), []string{})
		return &data.HttpResponse{
			Code:    code,
			Body:    body,
			Headers: headers,
			Error:   err,
		}, err

	case "post":
		code, body, headers, err := circuitbreaker.NewCircuitBreaker().HttpPost(req.Req.URL.String(), []byte{}, []string{})
		return &data.HttpResponse{
			Code:    code,
			Body:    body,
			Headers: headers,
			Error:   err,
		}, err

	default:
		panic("Implement '" + req.Req.Method + "'")

	}
	return nil, nil
}

func (a *attenuator) DoAsync(req *data.HttpRequest, callback AttenuatorCallback) error {
	return nil
}

func (a *attenuator) GetName() string {
	return a.name
}

func (a *attenuator) GetMaxHertz() float64 {
	return a.maxHertz
}

func (a *attenuator) GetTargetHertz() float64 {
	return a.targetHertz
}

func (a *attenuator) GetWorkers() int {
	return a.workers
}
