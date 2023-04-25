package attenuator

import (
	"http-attenuator/data"
)

// AttenuatorCallback is invoked by DoAsync() when the
// request is actually made
//
// TODO(john): async invocation
type AttenuatorCallback func(result AttenuatorResult) error

type Attenuator interface {
	DoSync(req *data.GatewayRequest) (*data.GatewayResponse, error)
	GetName() string
	GetMaxHertz() float64
	GetTargetHertz() float64
	GetWorkers() int
}

type AttenuatorResult interface {
	GetRequest() *data.GatewayRequest
	GetResponse() *data.GatewayResponse
	GetError() error
}

type attenuatorResult struct {
	Request  *data.GatewayRequest
	Response *data.GatewayResponse
	Error    error
}

func (a *attenuatorResult) GetRequest() *data.GatewayRequest {
	return a.Request
}

func (a *attenuatorResult) GetResponse() *data.GatewayResponse {
	return a.Response
}

func (a *attenuatorResult) GetError() error {
	return a.Error
}
