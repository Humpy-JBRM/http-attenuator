package data

import (
	"fmt"
	"io"
	"log"
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
	Body []byte `json:"body,omitempty"`
}

func (g *GatewayBase) GetUrl() *url.URL {
	return g.url
}

type GatewayRequest struct {
	GatewayBase

	Method string `json:"method"`
}

func NewGatewayRequestFromHttp(id string, req *http.Request) (*GatewayRequest, error) {
	var bodyBytes []byte
	var err error
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			e := fmt.Errorf("NewGatewayRequestFromHttp(%s): %s", req.URL.String(), err.Error())
			log.Println(e)
			return nil, e
		}
	}
	return NewGatewayRequest(
		id,
		req.Method,
		req.URL,
		req.Header,
		bodyBytes,
	), nil
}

func NewGatewayRequest(id string, method string, requestUrl *url.URL, headers http.Header, body []byte) *GatewayRequest {
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
	Id             string      `json:"id"`
	Headers        http.Header `json:"headers"`
	Body           []byte      `json:"body,omitempty"`
	StatusCode     int         `json:"status_code"`
	WhenMillis     int64       `json:"duration_millis"`
	DurationMillis int64       `json:"timestamp"`
	Error          error       `json:"error,omitempty"`
}

func NewGatewayResponse(id string, statusCode int, body []byte, headers http.Header, err error) *GatewayResponse {
	return &GatewayResponse{
		Id:         id,
		Headers:    headers,
		Body:       body,
		StatusCode: statusCode,
		Error:      err,
	}
}

func NewGatewayResponseFromHttp(id string, url *url.URL, resp *http.Response) (*GatewayResponse, error) {
	var bodyBytes []byte
	var err error
	if resp.Body != nil {
		bodyBytes, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			e := fmt.Errorf("NewGatewayRequestFromHttp(%s): %s", url.String(), err.Error())
			log.Println(e)
			return nil, e
		}
	}
	return NewGatewayResponse(
		id,
		resp.StatusCode,
		bodyBytes,
		resp.Header,
		err,
	), err
}
