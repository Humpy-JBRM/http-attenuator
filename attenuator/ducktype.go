package attenuator

import (
	"http-attenuator/data"
)

// AttenuatorCallback is invoked by DoAsync() when the
// request is actually made
type AttenuatorCallback func(result AttenuatorResult) error

type Attenuator interface {
	Stop()
	DoSync(req *data.HttpRequest) (*data.HttpResponse, error)
	DoAsync(req *data.HttpRequest, callback AttenuatorCallback) error
	GetName() string
	GetMaxHertz() float64
	GetTargetHertz() float64
	GetWorkers() int
}

type AttenuatorResult interface {
	GetRequest() *data.HttpRequest
	GetResponse() *data.HttpResponse
	GetError() error
}

type attenuatorResult struct {
	Request  *data.HttpRequest
	Response *data.HttpResponse
	Error    error
}

func (a *attenuatorResult) GetRequest() *data.HttpRequest {
	return a.Request
}

func (a *attenuatorResult) GetResponse() *data.HttpResponse {
	return a.Response
}

func (a *attenuatorResult) GetError() error {
	return a.Error
}
