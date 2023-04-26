package client

import (
	"fmt"
	"http-attenuator/data"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var httpClientRequests = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "http_client_requests",
		Help:      "The http_client requests, keyed by host, method and URI (without query string)",
	},
	[]string{"host", "method", "uri"},
)
var httpClientResponses = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "http_client_responses",
		Help:      "The http_client responses, keyed by host, method, URI (without query string) and status code",
	},
	[]string{"host", "method", "uri", "code"},
)
var httpClientRequestBytes = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "http_client_request_bytes",
		Help:      "The http_client bytes sent, keyed by host, method and URI (without query string)",
	},
	[]string{"host", "method", "uri"},
)
var httpClientResponseBytes = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "http_client_response_bytes",
		Help:      "The http_client bytes received, keyed by host, method and URI (without query string)",
	},
	[]string{"host", "method", "uri"},
)
var httpClientRequestsFailures = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "http_client_request_failures",
		Help:      "The http_client request failures, keyed by host, method and URI (without query string)",
	},
	[]string{"host", "method", "uri"},
)
var httpClientRequestsWaiting = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "http_client_requests_wait",
		Help:      "The http_client requests wait time in millis, keyed by host, method and URI (without query string)",
	},
	[]string{"host", "method", "uri"},
)
var httpClientRequestsLatency = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "http_client_requests_latency",
		Help:      "The http_client requests latency (round-trip-time) in millis, keyed by host, method and URI (without query string)",
	},
	[]string{"host", "method", "uri"},
)

type httpClientBuilder struct {
	impl HttpClientImpl
}

func NewHttpClientBuilder() HttpClientBuilder {
	return &httpClientBuilder{
		impl: HttpClientImpl{},
	}
}

func (cb *httpClientBuilder) Retries(retries int) HttpClientBuilder {
	cb.impl.Retries = retries
	return cb
}

func (cb *httpClientBuilder) TimeoutMillis(timeoutMillis int64) HttpClientBuilder {
	cb.impl.TimeoutMillis = timeoutMillis
	return cb
}

func (cb *httpClientBuilder) Success(fSuccess ...data.SuccessFunc) HttpClientBuilder {
	cb.impl.Success = append(cb.impl.Success, fSuccess...)
	return cb
}

func (cb *httpClientBuilder) Build() (HttpClient, error) {
	defensiveCopy := cb.impl
	if defensiveCopy.req != nil {
		// take a defensive clone of the request
		defensiveCopy.req = defensiveCopy.req.Clone(defensiveCopy.req.Context())
	}
	return &defensiveCopy, nil
}

func (cb *httpClientBuilder) Request(req *http.Request) HttpClientBuilder {
	cb.impl.req = req
	return cb
}

func (c *HttpClientImpl) Do(req *data.GatewayRequest) (*data.GatewayResponse, error) {
	var netClient = &http.Client{
		// TODO(john): CheckRedirect func(req *Request, via []*Request) error
		// TODO(john): Jar CookieJar
		// TODO(john): Transport
		Timeout: time.Duration(c.TimeoutMillis * int64(time.Millisecond)),
	}

	httpClientRequests.WithLabelValues(req.GetUrl().Host, req.GetRequest().Method, req.GetUrl().Path).Inc()
	httpClientRequestBytes.WithLabelValues(req.GetUrl().Host, req.GetRequest().Method, req.GetUrl().Path).Add(float64(len(req.Body)))
	nowMillis := time.Now().UTC().UnixMilli()
	request, err := http.NewRequest(strings.ToUpper(req.GetRequest().Method), req.GetUrl().String(), nil)
	if err != nil {
		resp := data.NewGatewayResponse(req.Id, http.StatusBadRequest, []byte{}, http.Header{}, err)
		resp.DurationMillis = (time.Now().UTC().UnixMilli() - nowMillis)
		httpClientRequestsFailures.WithLabelValues(req.GetUrl().Host, req.GetRequest().Method, req.GetUrl().Path).Inc()
		return resp, err
	}
	request.Header = req.Headers

	response, err := netClient.Do(request)
	if err != nil {
		resp, e := data.NewGatewayResponse(req.Id, http.StatusBadRequest, []byte{}, http.Header{}, err), err
		resp.DurationMillis = (time.Now().UTC().UnixMilli() - nowMillis)
		httpClientRequestsFailures.WithLabelValues(req.GetUrl().Host, req.GetRequest().Method, req.GetUrl().Path).Inc()
		return resp, e
	}
	httpClientRequestsLatency.WithLabelValues(req.GetUrl().Host, req.GetRequest().Method, req.GetUrl().Path).Add(float64(time.Now().UTC().UnixMilli() - nowMillis))
	if response == nil {
		e := fmt.Errorf("ERROR: %s: Got nil response from server", req.GetUrl().String())
		resp := data.NewGatewayResponse(req.Id, http.StatusBadRequest, []byte{}, http.Header{}, e)
		resp.DurationMillis = (time.Now().UTC().UnixMilli() - nowMillis)
		httpClientRequestsFailures.WithLabelValues(req.GetUrl().Host, req.GetRequest().Method, req.GetUrl().Path).Inc()
		return resp, e
	}

	responseBytes, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	httpClientResponseBytes.WithLabelValues(req.GetUrl().Host, req.GetRequest().Method, req.GetUrl().Path).Add(float64(len(responseBytes)))
	httpClientResponses.WithLabelValues(req.GetUrl().Host, req.GetRequest().Method, req.GetUrl().Path, fmt.Sprint(response.StatusCode)).Inc()

	resp := data.NewGatewayResponse(
		req.Id,
		response.StatusCode,
		responseBytes,
		response.Header,
		err,
	)
	resp.DurationMillis = (time.Now().UTC().UnixMilli() - nowMillis)
	return resp, err
}
