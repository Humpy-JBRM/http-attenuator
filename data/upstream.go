package data

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var upstreamRequests = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "upstream_requests",
		Help:      "The number of upstream requests, keyed by upstream/backend",
	},
	[]string{"tag", "upstream", "backend", "method"},
)
var upstreamErrors = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "upstream_errors",
		Help:      "The number of upstream errors, keyed by upstream/backend",
	},
	[]string{"tag", "upstream", "backend", "method"},
)
var upstreamLatency = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "upstream_latency",
		Help:      "The latency of upstream services, keyed by upstream/backend",
	},
	[]string{"tag", "upstream", "backend", "method"},
)
var upstreamResponses = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "upstream_responses",
		Help:      "The number of upstream responses, keyed by response code and upstream/backend",
	},
	[]string{"tag", "upstream", "backend", "method", "code"},
)

type Upstream interface {
	Handler
	ChooseBackend() UpstreamBackend
}

type UpstreamImpl struct {
	// this is backpatched
	serviceName string
	Cost        CostFromConfig                  `yaml:"cost" json:"cost"`
	Backends    map[string]*UpstreamBackendImpl `yaml:"backends" json:"backends"`
	Rule        string                          `yaml:"rule" json:"rule"`
	Pathology   string                          `yaml:"pathology" json:"pathology"`
	// These are backpatched
	backendCDF []HasCDF
	cost       Cost
	rng        *rand.Rand

	// this is backpatched to use the service broker
	HandlerFunc func(c *gin.Context)
}

func (u *UpstreamImpl) backpatch() error {
	u.rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	// Backpatch the cost
	u.cost = &CostImpl{
		coins: u.Cost,
	}

	// Backpatch the backends CDF
	totalWeight := 0
	for backendName, upstreamBackend := range u.Backends {
		upstreamBackend.backendName = backendName
		totalWeight += upstreamBackend.Weight
	}
	u.backendCDF = make([]HasCDF, 0)
	for _, upstreamBackend := range u.Backends {
		upstreamBackend.cdf = float64(upstreamBackend.Weight) / float64(totalWeight)
		u.backendCDF = append(u.backendCDF, upstreamBackend)
	}

	return nil
}

func (u *UpstreamImpl) GetName() string {
	return u.serviceName
}

func (u *UpstreamImpl) Handle(c *gin.Context) {
	// Make sure the varz get updated
	nowMillis := time.Now().UTC().UnixMilli()
	var backend UpstreamBackend
	defer func() {
		var backendName string
		if backend != nil {
			backendName = backend.GetName()
		}
		upstreamLatency.WithLabelValues(
			c.Request.Header.Get(HEADER_X_FAULTMONKEY_TAG),
			u.GetName(), // upstream
			backendName, // backend
			c.Request.Method,
		).Add(float64(time.Now().UTC().UnixMilli() - nowMillis))
	}()

	// Select a backend
	backend = u.ChooseBackend()
	if backend == nil {
		err := fmt.Errorf("Handle(%s): No backend for '%s.%s'", c.Request.URL, u.serviceName, u.GetName())
		log.Println(err)
		upstreamErrors.WithLabelValues(
			c.Request.Header.Get(HEADER_X_FAULTMONKEY_TAG),
			u.GetName(), // upstream
			"",          // backend
			c.Request.Method,
		).Inc()
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	// Update the varz
	upstreamRequests.WithLabelValues(
		c.Request.Header.Get(HEADER_X_FAULTMONKEY_TAG),
		u.GetName(),
		backend.GetName(), // backend
		c.Request.Method,
	).Inc()

	// Delegate to the handler function
	backend.Handle(c)
}

func (u *UpstreamImpl) ChooseBackend() UpstreamBackend {
	backend := Choose(u.Rule, u.backendCDF, u.rng)
	if backend == nil {
		return nil
	}

	return backend.(UpstreamBackend)
}

type UpstreamBackend interface {
	HasCDF
	Handler
	GetCost() Cost
	GetURL() *url.URL
}

type UpstreamBackendImplFromConfig map[string]*UpstreamBackendImpl

type UpstreamBackendImpl struct {
	// this is backpatched
	backendName string
	Impl        string `yaml:"impl" json:"impl"`
	Url         string `yaml:"url" json:"url"`
	Weight      int    `yaml:"weight" json:"weight"`
	Pathology   string `yaml:"pathology" json:"pathology"`

	// This is used to override the default cost for this upstream.
	//
	// It allows us to (for instance) implement some kind of
	// ChooseCheapest()
	Cost Cost `yaml:"cost,omitempty" json:"impl,omitempty"`

	// This is backpatched
	cdf float64
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

func (u *UpstreamBackendImpl) Handle(c *gin.Context) {
	request := *c.Request
	request.URL = u.GetURL()
	request.Host = u.GetURL().Host

	//http: Request.RequestURI can't be set in client requests.
	//http://golang.org/src/pkg/net/http/client.go
	request.RequestURI = ""

	// Make the request
	//
	// TODO(john): put it through the attenuator / circuit breaker etc
	client := http.Client{}
	resp, err := client.Do(&request)
	if err != nil {
		c.Writer.Header().Add(HEADER_X_ATTENUATOR_ERROR, err.Error())
		upstreamResponses.WithLabelValues(
			c.Request.Header.Get(HEADER_X_FAULTMONKEY_TAG),
			u.backendName,
			u.GetName(),
			c.Request.Method,
		).Inc()
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}

	defer resp.Body.Close()

	// Send the status
	upstreamResponses.WithLabelValues(
		c.Request.Header.Get(HEADER_X_FAULTMONKEY_TAG),
		u.backendName,
		u.GetName(),
		c.Request.Method,
		fmt.Sprint(resp.StatusCode),
	).Inc()
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
