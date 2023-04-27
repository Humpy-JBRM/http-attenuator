package data

import (
	"bytes"
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

	req *http.Request
}

func (r *GatewayRequest) GetRequest() *http.Request {
	return r.req
}

func NewGatewayRequestFromHttp(id string, req *http.Request) (*GatewayRequest, error) {
	var bodyBytes []byte
	var err error
	if req.Body != nil {
		// Clone the body so we don't mess up the io.Reader
		cloned := req.Clone(req.Context())
		bodyBytes, err = io.ReadAll(cloned.Body)
		cloned.Body.Close()
		if err != nil {
			e := fmt.Errorf("NewGatewayRequestFromHttp(%s): %s", req.URL.String(), err.Error())
			log.Println(e)
			return nil, e
		}
	}
	gwr := &GatewayRequest{
		GatewayBase: GatewayBase{
			Id:         id,
			url:        req.URL,
			DisplayUrl: req.URL.String(),
			Headers:    req.Header,
			Body:       bodyBytes,
		},
	}
	return gwr, nil
}

func NewGatewayRequest(id string, method string, requestUrl *url.URL, headers http.Header, body []byte) (*GatewayRequest, error) {
	idToUse := id
	if id == "" {
		idToUse = uuid.NewString()
	}

	req, err := http.NewRequest(
		method,
		requestUrl.String(),
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header = headers
	gwr := &GatewayRequest{
		GatewayBase: GatewayBase{
			Id:         idToUse,
			url:        requestUrl,
			DisplayUrl: requestUrl.String(),
			Headers:    headers,
			Body:       body,
		},
		req: req,
	}

	return gwr, nil
}

type GatewayResponse struct {
	GatewayBase
	StatusCode     int   `json:"status_code"`
	WhenMillis     int64 `json:"duration_millis"`
	DurationMillis int64 `json:"timestamp"`
	Error          error `json:"error,omitempty"`
	resp           *http.Response
}

func (r *GatewayResponse) GetResponse() *http.Response {
	return r.resp
}

func NewGatewayResponse(id string, statusCode int, body []byte, headers http.Header, err error) *GatewayResponse {
	return &GatewayResponse{
		GatewayBase: GatewayBase{
			Id:      id,
			Headers: headers,
			Body:    body,
		},
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
	gwr := &GatewayResponse{
		GatewayBase: GatewayBase{
			Id:         id,
			url:        resp.Request.URL,
			DisplayUrl: resp.Request.URL.String(),
			Headers:    resp.Header,
			Body:       bodyBytes,
		},
		StatusCode: resp.StatusCode,
	}

	return gwr, err
}
