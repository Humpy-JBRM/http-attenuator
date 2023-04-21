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

	// The URL
	Url url.URL `json:"url"`

	// Headers
	Headers http.Header `json:"headers"`

	// Request body (if any)
	Body *[]byte `json:"body,omitempty"`
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
			Id:      idToUse,
			Url:     *requestUrl,
			Headers: headers,
			Body:    body,
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
