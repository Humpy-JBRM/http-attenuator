package client

import (
	"context"
	"fmt"
	"http-attenuator/data"
	"net/http"
)

// HttpClient handles all of the retry logic etc
//
// It encapsulates all of:
//
//   - retry logic
//
//   - TLS config
//
//   - rate/limiting and attenuation
type HttpClient interface {
	Do(ctx context.Context, req *data.GatewayRequest) (*data.GatewayResponse, error)
}

type HttpClientImpl struct {
	// The attenuator for this client.
	//
	// This gets looked up in the attenuator registry
	//
	// TODO(john): make this hierarchical (so we can [e.g.] default to a global or service level)
	AttenuatorName string `json:"attenuator"`

	// The maximum number of retries
	Retries int `json:"retries"`

	// The HTTP-level timeout in milliseconds
	TimeoutMillis int64 `json:"timeout_millis"`

	// Functions to execute to determine whether or not a request
	// was successful
	Success []data.SuccessFunc `json:"-"`

	// If we are recording requests, this is where we do it
	RecordRequestRoot  string `json:"record_request_root"`
	RecordResponseRoot string `json:"record_response_root"`

	// Underlying http request
	req *http.Request

	// attenuator instance
	// If this is nil, there is no attenuation
	attenuator Attenuator
}

type HttpClientBuilder interface {
	Attenuator(attenuator Attenuator) HttpClientBuilder
	Retries(retries int) HttpClientBuilder
	TimeoutMillis(timeoutMillis int64) HttpClientBuilder
	Success(fSuccess ...data.SuccessFunc) HttpClientBuilder
	RecordRequest(recordRequestRoot string) HttpClientBuilder
	RecordResponse(recordResponseRoot string) HttpClientBuilder
	Build() (HttpClient, error)
}

type ErrWaitTimeout struct {
	msg string
}

func (e *ErrWaitTimeout) Error() string {
	return e.msg
}

func NewErrWaitTimeout(msg string) *ErrWaitTimeout {
	return &ErrWaitTimeout{
		msg: msg,
	}
}

type Attenuator interface {
	fmt.Stringer
	//DoSync(req *data.GatewayRequest) (*data.GatewayResponse, error)
	GetName() string
	GetMaxHertz() float64
	GetMaxInflight() int
	WaitForGreen(ctx context.Context, cancelFunc context.CancelFunc) error
}
