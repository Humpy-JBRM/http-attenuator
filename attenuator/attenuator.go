package attenuator

import (
	"fmt"
	client "http-attenuator/client"
	"http-attenuator/data"
	config "http-attenuator/facade/config"
	trafficlight "http-attenuator/facade/trafficlight"
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
	pulse := trafficlight.GetPulse(name)
	var err error
	if pulse == nil {
		// TODO(john): hook this up and encapsulate it behind the pulse
		// factory, so we can use redis
		pulse, err = trafficlight.NewPulse(name, workers, maxHertz, targetHertz)
		if err != nil {
			return nil, err
		}
	}
	if pulse == nil {
		return nil, fmt.Errorf("Unable to get pulse '%s'", name)
	}

	trafficlight.RegisterTrafficLight(
		&trafficlight.TrafficLightImpl{
			Name:  name,
			Pulse: pulse,
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
	trafficlight.WaitForGreen(a.name, 1)
	attenuatedRequestsWaiting.WithLabelValues(req.GetUrl().Host, req.Method, req.GetUrl().Path).Add(float64(time.Now().UTC().UnixMilli() - nowMillis))
	attenuatedRequests.WithLabelValues(req.GetUrl().Host, req.Method, req.GetUrl().Path).Inc()

	// Do the request
	var err error
	cb, err := client.NewHttpClientBuilder().
		Retries(0).
		TimeoutMillis(10000).
		Build()
	if err != nil {
		// out of switch
		attenuatedRequestsFailures.WithLabelValues(req.GetUrl().Host, req.Method, req.GetUrl().Path).Inc()
		return nil, err
	}

	resp, err := cb.Do(req)
	attenuatedRequestsLatency.WithLabelValues(req.GetUrl().Host, req.Method, req.GetUrl().Path).Add(float64(time.Now().UTC().UnixMilli() - nowMillis))
	if err != nil {
		attenuatedRequestsFailures.WithLabelValues(req.GetUrl().Host, req.Method, req.GetUrl().Path).Inc()
	}
	return resp, err
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
