package circuitbreaker

import (
	"http-attenuator/data"
	"net/http"
)

// CircuitBreaker handles all of the retry logic etc
type CircuitBreaker interface {
	HttpGet(*data.GatewayRequest) (int, []byte, http.Header, error)
	HttpPost(*data.GatewayRequest) (int, []byte, http.Header, error)
}

type CircuitBreakerImpl struct {
	// The name of the traffic light we use to attenuate this
	TrafficLight string `json:"traffic_light"`

	// The maximum number of retries
	Retries int `json:"retries"`

	// The HTTP-level timeout in milliseconds
	TimeoutMillis int64 `json:"timeout_millis"`

	// Functions to execute to determine whether or not a request
	// was successful
	Success []data.SuccessFunc `json:"-"`
}

type CircuitBreakerBuilder interface {
	TrafficLight(trafficLight string) CircuitBreakerBuilder
	Retries(retries int) CircuitBreakerBuilder
	TimeoutMillis(timeoutMillis int64) CircuitBreakerBuilder
	Success(fSuccess ...data.SuccessFunc) CircuitBreakerBuilder
	Build() (CircuitBreaker, error)
}
