package data

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

type UpstreamBackend interface {
	HasCDF
	Handler
	GetCost() Cost
	GetURL() *url.URL

	// Dynamic config
	Disable() error
	Enable() error

	// For starting and stopping the healthcheck.
	// TODO(john): these should be a separate interface
	Start()
	Stop()
}

type UpstreamBackendImplFromConfig map[string]*UpstreamBackendImpl

type BackendState int

const (
	UNKNOWN BackendState = iota
	HEALTHY
	UNHEALTHY
	DISABLED
)

func (s BackendState) String() string {
	switch s {
	case HEALTHY:
		return "HEALTHY"

	case UNHEALTHY:
		return "UNHEALTHY"

	case DISABLED:
		return "DISABLED"

	default:
		return "UNKNOWN"
	}
}

type UpstreamBackendImpl struct {
	// this is backpatched
	backendName string
	Impl        string        `yaml:"impl" json:"impl"`
	Url         string        `yaml:"url" json:"url"`
	Weight      int           `yaml:"weight" json:"weight"`
	Pathology   string        `yaml:"pathology" json:"pathology"`
	Recorder    *RecorderImpl `yaml:"recorder" json:"recorder"`

	// This is used to override the default cost for this upstream.
	//
	// It allows us to (for instance) implement some kind of
	// ChooseCheapest()
	Cost Cost `yaml:"cost,omitempty" json:"impl,omitempty"`

	// This is backpatched
	cdf float64

	// State of this backend
	State BackendState

	// The healthcheck spec and implementation
	Healthcheck         string `yaml:"healthcheck" json:"healthcheck"`
	healthCheckImpl     EVT
	healthCheckInterval *time.Duration
	healthcheckResult   EVTResult
	healthcheckStop     chan interface{}
}

func (u *UpstreamBackendImpl) GetName() string {
	return u.backendName
}

func (u *UpstreamBackendImpl) CDF() float64 {
	return u.cdf
}

func (u *UpstreamBackendImpl) SetCDF(cdf float64) {
	u.cdf = cdf
}

func (u *UpstreamBackendImpl) GetWeight() int {
	return u.Weight
}

func (u *UpstreamBackendImpl) GetCost() Cost {
	return u.Cost
}

func (u *UpstreamBackendImpl) GetURL() *url.URL {
	parsedUrl, _ := url.Parse(u.Url)
	return parsedUrl
}

func (u *UpstreamBackendImpl) Disable() error {
	return nil
}

func (u *UpstreamBackendImpl) Enable() error {
	return nil
}

func (u *UpstreamBackendImpl) SetHealthcheck(healthcheck EVT) {
	u.healthCheckImpl = healthcheck

	// TODO(john): get this from config
	d := time.Duration(time.Second * 10)
	u.healthCheckInterval = &d
}

func (u *UpstreamBackendImpl) Start() {
	u.healthcheckStop = make(chan interface{})
	go u.healthcheckWorker()
}

func (u *UpstreamBackendImpl) Stop() {
	u.healthcheckStop <- true
}

func (u *UpstreamBackendImpl) healthcheckWorker() {
	if u.healthCheckImpl == nil {
		return
	}
	log.Printf("%s: starting healthcheck '%s'", u.GetName(), u.Healthcheck)
	defer func() {
		log.Printf("%s: terminating healthcheck '%s'", u.GetName(), u.Healthcheck)
	}()
	for {
		select {
		case <-u.healthcheckStop:
			return
		default:
			// Health check is still active
		}
		if u.State != DISABLED {
			u.healthcheckResult = u.healthCheckImpl.Run()
			if u.healthcheckResult.Error() != nil {
				u.State = UNHEALTHY
			}
		}
		time.Sleep(*u.healthCheckInterval)
	}
}

func (u *UpstreamBackendImpl) Handle(c *gin.Context) {
	request := *c.Request
	request.URL = u.GetURL()
	request.Host = u.GetURL().Host

	//http: Request.RequestURI can't be set in client requests.
	//http://golang.org/src/pkg/net/http/client.go
	request.RequestURI = ""
	now := time.Now().UTC().UnixMilli()

	if u.Recorder != nil {
		// TODO(john): this is ugly and inefficient.  Implement something more elegant
		requestBody, _ := io.ReadAll(request.Body)
		request.Body.Close()
		request.Body = io.NopCloser(bytes.NewReader(requestBody))
		gwr, _ := NewGatewayRequest(
			"",
			request.Method,
			request.URL,
			request.Header,
			requestBody,
		)
		gwr.WhenMillis = now
		u.Recorder.SaveRequest(gwr)
	}

	requestHeaders := make(http.Header)
	requestHeaders.Add(HEADER_X_FAULTMONKEY_API_CUSTOMER, request.Header.Get(HEADER_X_FAULTMONKEY_API_CUSTOMER))
	requestHeaders.Add(HEADER_X_REQUEST_ID, request.Header.Get(HEADER_X_REQUEST_ID))
	requestHeaders.Add(HEADER_X_FAULTMONKEY_BACKEND, request.Header.Get(HEADER_X_FAULTMONKEY_BACKEND))
	requestHeaders.Add(HEADER_X_FAULTMONKEY_UPSTREAM, request.Header.Get(HEADER_X_FAULTMONKEY_UPSTREAM))
	requestHeaders.Add(HEADER_X_FAULTMONKEY_TAG, request.Header.Get(HEADER_X_FAULTMONKEY_TAG))

	// Make the request
	//
	// TODO(john): put it through the attenuator / circuit breaker etc
	client := http.Client{}
	resp, err := client.Do(&request)
	if err != nil {
		log.Printf("%s: %s", request.URL.String(), err.Error())
		c.Writer.Header().Add(HEADER_X_FAULTMONKEY_ERROR, err.Error())
		upstreamResponses.WithLabelValues(
			c.Request.Header.Get(HEADER_X_FAULTMONKEY_TAG),
			u.GetName(),
			c.Request.Header.Get(HEADER_X_FAULTMONKEY_BACKEND),
			c.Request.Method,
			fmt.Sprint(http.StatusBadGateway),
		).Inc()

		errorClass := GetErrorClassifier().Classify(err)
		upstreamErrors.WithLabelValues(
			c.Request.Header.Get(HEADER_X_FAULTMONKEY_TAG),
			u.GetName(),
			c.Request.Header.Get(HEADER_X_FAULTMONKEY_BACKEND),
			c.Request.Method,
			errorClass,
		).Inc()
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}

	// Propagatethe request headers into the response
	for header, val := range requestHeaders {
		if resp.Header.Get(header) == "" {
			resp.Header.Add(header, val[0])
		}
	}

	latency := time.Now().UTC().UnixMilli() - now
	resp.Header.Add(HEADER_X_FAULTMONKEY_BACKEND_LATENCY, fmt.Sprint(latency))
	if u.Recorder != nil {
		// TODO(john): this is ugly and inefficient.  Implement something more elegant
		responseBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(responseBody))
		gwr := NewGatewayResponse(
			request.Header.Get(HEADER_X_REQUEST_ID),
			resp.StatusCode,
			responseBody,
			resp.Header,
			nil,
		)
		gwr.WhenMillis = now
		gwr.DurationMillis = latency
		u.Recorder.SaveResponse(gwr)
	}
	defer resp.Body.Close()

	// Send the status
	upstreamLatency.WithLabelValues(
		c.Request.Header.Get(HEADER_X_FAULTMONKEY_TAG),
		u.GetName(),
		c.Request.Header.Get(HEADER_X_FAULTMONKEY_BACKEND),
		c.Request.Method,
		fmt.Sprint(resp.StatusCode),
	).Add(float64(latency))
	upstreamResponses.WithLabelValues(
		c.Request.Header.Get(HEADER_X_FAULTMONKEY_TAG),
		u.GetName(),
		c.Request.Header.Get(HEADER_X_FAULTMONKEY_BACKEND),
		c.Request.Method,
		fmt.Sprint(resp.StatusCode),
	).Inc()
	if resp.StatusCode >= 400 {
		upstreamErrors.WithLabelValues(
			c.Request.Header.Get(HEADER_X_FAULTMONKEY_TAG),
			u.GetName(),
			c.Request.Header.Get(HEADER_X_FAULTMONKEY_BACKEND),
			c.Request.Method,
			fmt.Sprint(resp.StatusCode),
		).Inc()
	}
	c.Status(resp.StatusCode)

	// Send the headers
	for h, v := range resp.Header {
		for _, headerVal := range v {
			c.Writer.Header().Add(h, headerVal)
		}
	}

	// Send the body
	io.Copy(c.Writer, resp.Body)
}
