package circuitbreaker

import (
	"fmt"
	"http-attenuator/data"
	"http-attenuator/util"
	"net/http"
)

type circuitBreakerBuilder struct {
	impl CircuitBreakerImpl
}

func NewCircuitBreakerBuilder() CircuitBreakerBuilder {
	return &circuitBreakerBuilder{
		impl: CircuitBreakerImpl{},
	}
}

func (cb *circuitBreakerBuilder) TrafficLight(trafficLight string) CircuitBreakerBuilder {
	cb.impl.TrafficLight = trafficLight
	return cb
}

func (cb *circuitBreakerBuilder) Retries(retries int) CircuitBreakerBuilder {
	cb.impl.Retries = retries
	return cb
}

func (cb *circuitBreakerBuilder) TimeoutMillis(timeoutMillis int64) CircuitBreakerBuilder {
	cb.impl.TimeoutMillis = timeoutMillis
	return cb
}

func (cb *circuitBreakerBuilder) Success(fSuccess ...data.SuccessFunc) CircuitBreakerBuilder {
	cb.impl.Success = append(cb.impl.Success, fSuccess...)
	return cb
}

func (cb *circuitBreakerBuilder) Build() (CircuitBreaker, error) {
	defensiveCopy := cb.impl
	return &defensiveCopy, nil
}

func (c *CircuitBreakerImpl) HttpGet(req *data.GatewayRequest) (int, []byte, http.Header, error) {
	headers := make([]string, 0)
	for k, v := range req.Headers {
		if len(v) > 0 {
			headers = append(headers, fmt.Sprintf("%s: %s", k, v))
		}
	}
	return util.HttpGet(req.Url.String(), req.Headers)
}

func (c *CircuitBreakerImpl) HttpPost(req *data.GatewayRequest) (int, []byte, http.Header, error) {
	headers := make([]string, 0)
	var body []byte
	if req.Body != nil {
		body = *req.Body
	}
	for k, v := range req.Headers {
		if len(v) > 0 {
			headers = append(headers, fmt.Sprintf("%s: %s", k, v))
		}
	}
	return util.HttpPost(req.Url.String(), body, req.Headers)
}
