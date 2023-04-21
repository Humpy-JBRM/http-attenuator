package circuitbreaker

import (
	"http-attenuator/data"
	"http-attenuator/util"
)

// CircuitBreaker handles all of the retry logic etc
type CircuitBreaker interface {
	HttpGet(theUrl string, header []string, success ...data.SuccessFunc) (int, []byte, map[string]string, error)
	HttpPost(theUrl string, payload []byte, header []string, success ...data.SuccessFunc) (int, []byte, map[string]string, error)
}

type circuitBreaker struct {
}

func NewCircuitBreaker() CircuitBreaker {
	return &circuitBreaker{}
}

func (c *circuitBreaker) HttpGet(theUrl string, header []string, success ...data.SuccessFunc) (int, []byte, map[string]string, error) {
	return util.HttpGet(theUrl, header...)
}

func (c *circuitBreaker) HttpPost(theUrl string, payload []byte, header []string, success ...data.SuccessFunc) (int, []byte, map[string]string, error) {
	return util.HttpPost(theUrl, payload, header...)
}
