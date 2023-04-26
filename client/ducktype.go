package client

import (
	"http-attenuator/data"
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
	Do(*data.GatewayRequest) (*data.GatewayResponse, error)
	Get(*data.GatewayRequest) (*data.GatewayResponse, error)
	Post(*data.GatewayRequest) (*data.GatewayResponse, error)
}

type HttpClientImpl struct {
	// The maximum number of retries
	Retries int `json:"retries"`

	// The HTTP-level timeout in milliseconds
	TimeoutMillis int64 `json:"timeout_millis"`

	// Functions to execute to determine whether or not a request
	// was successful
	Success []data.SuccessFunc `json:"-"`
}

type HttpClientBuilder interface {
	Retries(retries int) HttpClientBuilder
	TimeoutMillis(timeoutMillis int64) HttpClientBuilder
	Success(fSuccess ...data.SuccessFunc) HttpClientBuilder
	Build() (HttpClient, error)
}
