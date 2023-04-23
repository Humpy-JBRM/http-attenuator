package data

import (
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

// SuccessFunc returns two bools:
//
//	bool	true == Success
//	bool	true == retry (only applies if success is false)
type SuccessFunc func(resp *http.Response) (bool, bool)

type GatewayBase struct {
	// For tracing
	Id string `json:"id"`

	// For tracing / replaying / organising
	WhenMillis int64 `json:"timestamp"`

	// The URL
	// Make this private so that json (un)marshall works correctly and
	// does not complain about trying to marshal / unmarshall a function
	url *url.URL `json:"-"`

	DisplayUrl string `json:"url"`

	// Headers
	Headers http.Header `json:"headers"`

	// Request body (if any)
	Body *[]byte `json:"body,omitempty"`
}

func (g *GatewayBase) GetUrl() *url.URL {
	return g.url
}

type GatewayRequest struct {
	GatewayBase

	Method string `json:"method"`
}

func NewGatewayRequest(id string, method string, requestUrl *url.URL, headers http.Header, body *[]byte) *GatewayRequest {
	idToUse := id
	if id == "" {
		idToUse = uuid.NewString()
	}

	gwr := &GatewayRequest{
		GatewayBase: GatewayBase{
			Id:         idToUse,
			url:        requestUrl,
			DisplayUrl: requestUrl.String(),
			Headers:    headers,
			Body:       body,
		},
		Method: method,
	}

	return gwr
}

type GatewayResponse struct {
	GatewayBase
	StatusCode     int   `json:"status_code"`
	DurationMillis int64 `json:"duration_millis"`
}

func NewGatewayResponse() *GatewayResponse {
	return &GatewayResponse{}
}
