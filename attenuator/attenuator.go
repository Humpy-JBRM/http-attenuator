package attenuator

import (
	"fmt"
	"http-attenuator/circuitbreaker"
	"http-attenuator/data"
	config "http-attenuator/facade/config"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/spf13/viper"
)

var attenuatedRequests = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "attenuated_requests",
		Help:      "The attenuated requests, keyed by host, method and URI (without query string)",
	},
	[]string{"host", "method", "uri"},
)
var attenuatedRequestsFailures = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "attenuated_request_failures",
		Help:      "The attenuated request failures, keyed by host, method and URI (without query string)",
	},
	[]string{"host", "method", "uri"},
)
var attenuatedRequestsWaiting = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "attenuated_requests_wait",
		Help:      "The attenuated requests wait time in millis, keyed by host, method and URI (without query string)",
	},
	[]string{"host", "method", "uri"},
)
var attenuatedRequestsLatency = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "attenuated_requests_latency",
		Help:      "The attenuated requests latency (round-trip-time) in millis, keyed by host, method and URI (without query string)",
	},
	[]string{"host", "method", "uri"},
)

type attenuator struct {
	name         string
	maxHertz     float64
	targetHertz  float64
	workers      int
	requestQueue chan (*data.GatewayRequest)
	workerCount  int64
	stopped      bool
	trafficLight string
}

func NewAttenuator(name string, maxHertz float64, targetHertz float64, workers int) (Attenuator, error) {
	pulse := GetPulse(name)
	var err error
	if pulse == nil {
		pulse, err = NewPulse(name, workers, maxHertz, targetHertz)
		if err != nil {
			return nil, err
		}
	}
	if pulse == nil {
		return nil, fmt.Errorf("Unable to get pulse '%s'", name)
	}

	RegisterTrafficLight(
		&TrafficLightImpl{
			Name:  name,
			pulse: pulse,
		})

	queueSize, _ := config.Config().GetInt(data.CONF_ATTENUATOR_QUEUESIZE)
	if queueSize <= 0 {
		return nil, fmt.Errorf("cannot have an attenuator queue size of %d.  Is '%s' set correctly?", queueSize, data.CONF_ATTENUATOR_QUEUESIZE)
	}

	a := &attenuator{
		name:     name,
		maxHertz: maxHertz,

		// targetHertz is not implement yet
		// it is used for auto-scaling up and down
		targetHertz:  0,
		workers:      workers,
		requestQueue: make(chan (*data.GatewayRequest), viper.GetInt(data.CONF_ATTENUATOR_QUEUESIZE)),
		trafficLight: name,
	}

	return a, nil
}

func (a *attenuator) DoSync(req *data.GatewayRequest) (*data.GatewayResponse, error) {
	// wait for green light
	nowMillis := time.Now().UTC().UnixMilli()
	WaitForGreen(a.name, 1)
	attenuatedRequestsWaiting.WithLabelValues(req.Url.Host, req.Method, req.Url.Path).Add(float64(time.Now().UTC().UnixMilli() - nowMillis))
	attenuatedRequests.WithLabelValues(req.Url.Host, req.Method, req.Url.Path).Inc()

	// Do the request
	var err error
	switch strings.ToLower(req.Method) {
	case "get":
		cb, err := circuitbreaker.NewCircuitBreakerBuilder().
			TrafficLight("").
			Retries(0).
			TimeoutMillis(10000).
			Build()
		if err != nil {
			// out of switch
			break
		}

		nowMillis := time.Now().UTC().UnixMilli()
		code, body, headers, err := cb.HttpGet(req)
		resp := &data.GatewayResponse{
			GatewayBase:    req.GatewayBase,
			StatusCode:     code,
			DurationMillis: (time.Now().UTC().UnixMilli() - nowMillis),
		}
		resp.Body = &body
		resp.Headers = headers
		attenuatedRequestsLatency.WithLabelValues(req.Url.Host, req.Method, req.Url.Path).Add(float64(time.Now().UTC().UnixMilli() - nowMillis))
		return resp, err

	case "post":
		cb, err := circuitbreaker.NewCircuitBreakerBuilder().
			TrafficLight("").
			Retries(0).
			TimeoutMillis(10000).
			Build()
		if err != nil {
			// out of switch
			break
		}

		nowMillis := time.Now().UTC().UnixMilli()
		code, body, headers, err := cb.HttpPost(req)
		resp := &data.GatewayResponse{
			GatewayBase:    req.GatewayBase,
			StatusCode:     code,
			DurationMillis: (time.Now().UTC().UnixMilli() - nowMillis),
		}
		resp.Body = &body
		resp.Headers = headers
		attenuatedRequestsLatency.WithLabelValues(req.Url.Host, req.Method, req.Url.Path).Add(float64(time.Now().UTC().UnixMilli() - nowMillis))
		return resp, err

	default:
		panic(fmt.Sprintf("Implement %s %s", strings.ToUpper(req.Method), req.Url.String()))

	}
	if err != nil {
		attenuatedRequestsFailures.WithLabelValues(req.Url.Host, req.Method, req.Url.Path).Inc()
	}
	return nil, err
}

func (a *attenuator) DoAsync(req *data.GatewayRequest, callback AttenuatorCallback) error {
	panic("Async not implemented in MVP")
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
