package data

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	[]string{"tag", "upstream", "backend", "method", "code"},
)
var upstreamLatency = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "upstream_latency",
		Help:      "The latency of upstream services, keyed by upstream/backend",
	},
	[]string{"tag", "upstream", "backend", "method", "code"},
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
	ChooseBackend(preferredBackend string) UpstreamBackend
}

type UpstreamImpl struct {
	// this is backpatched
	serviceName string
	Cost        CostFromConfig                  `yaml:"cost" json:"cost"`
	Backends    map[string]*UpstreamBackendImpl `yaml:"backends" json:"backends"`
	Rule        string                          `yaml:"rule" json:"rule"`
	Pathology   string                          `yaml:"pathology" json:"pathology"`
	Recorder    *RecorderImpl                   `yaml:"recorder" json:"recorder"`

	// These are backpatched
	backendCDF []HasCDF
	cost       Cost
	rng        *rand.Rand

	// this is backpatched to use the service broker
	HandlerFunc func(c *gin.Context)
}

func (u *UpstreamImpl) Backpatch() error {
	u.rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	// Backpatch the cost
	u.cost = &CostImpl{
		coins: u.Cost,
	}

	// Kick off the request / response recorder
	var err error
	if u.Recorder != nil {
		err = u.Recorder.Backpatch()
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

	// Backends can have their own recorder.
	// If they don't have one specified, it uses the one from the upstream parent
	for _, upstreamBackend := range u.Backends {
		if upstreamBackend.Recorder == nil {
			upstreamBackend.Recorder = u.Recorder
		} else {
			upstreamBackend.Recorder.Backpatch()
		}
	}

	return err
}

func (u *UpstreamImpl) GetName() string {
	return u.serviceName
}

func (u *UpstreamImpl) Handle(c *gin.Context) {
	// Make sure the varz get updated
	nowMillis := time.Now().UTC().UnixMilli()
	var backend UpstreamBackend
	var responseCodeAsString string
	defer func() {
		upstreamLatency.WithLabelValues(
			c.Request.Header.Get(HEADER_X_FAULTMONKEY_TAG),
			u.GetName(),
			c.Request.Header.Get(HEADER_X_FAULTMONKEY_BACKEND),
			c.Request.Method,
			responseCodeAsString,
		).Add(float64(time.Now().UTC().UnixMilli() - nowMillis))
	}()

	// Select a backend
	backend = u.ChooseBackend(c.GetHeader(HEADER_X_FAULTMONKEY_BACKEND))
	if backend == nil {
		err := fmt.Errorf("Handle(%s): No backend for '%s.%s'", c.Request.URL, u.serviceName, u.GetName())
		log.Println(err)
		responseCodeAsString = fmt.Sprint(http.StatusBadGateway)
		upstreamErrors.WithLabelValues(
			c.Request.Header.Get(HEADER_X_FAULTMONKEY_TAG),
			u.GetName(),
			c.Request.Header.Get(HEADER_X_FAULTMONKEY_BACKEND),
			c.Request.Method,
			responseCodeAsString,
		).Inc()
		c.AbortWithError(http.StatusBadGateway, err)

		return
	}
	if c.GetHeader(HEADER_X_REQUEST_ID) == "" {
		c.Request.Header.Add(HEADER_X_REQUEST_ID, uuid.NewString())
	}
	if c.GetHeader(HEADER_X_FAULTMONKEY_BACKEND) == "" {
		c.Request.Header.Add(HEADER_X_FAULTMONKEY_BACKEND, backend.GetName())
	}
	if c.GetHeader(HEADER_X_FAULTMONKEY_UPSTREAM) == "" {
		c.Request.Header.Add(HEADER_X_FAULTMONKEY_UPSTREAM, u.GetName())
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

func (u *UpstreamImpl) ChooseBackend(preferredBackend string) UpstreamBackend {
	if preferredBackend != "" {
		// The caller has specified that there is a specific backend they want
		// to use
		return u.Backends[preferredBackend]
	}

	backend := Choose(u.Rule, u.backendCDF, u.rng)
	if backend == nil {
		return nil
	}

	return backend.(UpstreamBackend)
}
