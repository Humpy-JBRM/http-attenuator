package api

import (
	"fmt"
	"http-attenuator/broker"
	"http-attenuator/data"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var brokerRequests = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "broker_requests",
		Help:      "The number of broker requests, keyed by service",
	},
	[]string{"tag", "service", "backend", "method"},
)
var brokerRequestsLatency = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "broker_requests_latency",
		Help:      "The latency of broker requests, keyed by service",
	},
	[]string{"tag", "service", "backend", "method"},
)
var brokerResponses = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "broker_responses",
		Help:      "The number of broker responses, keyed by response code and service",
	},
	[]string{"tag", "service", "backend", "method", "code"},
)
var brokerCost = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "broker_cost",
		Help:      "The cost of broker responses, keyed by service, backend and customer",
	},
	[]string{"tag", "service", "backend", "customer"},
)

func BrokerHandler(c *gin.Context) {
	// Extract the service from the URL
	serviceAndUri := c.Param("serviceAndUri")
	if serviceAndUri == "" {
		brokerRequests.WithLabelValues(
			c.Request.Header.Get(data.HEADER_X_MIGALOO_TAG),
			"",
			"",
			c.Request.Method,
		).Inc()
		brokerResponses.WithLabelValues(
			c.Request.Header.Get(data.HEADER_X_MIGALOO_TAG),
			"",
			"",
			c.Request.Method,
			fmt.Sprint(http.StatusNotFound),
		).Inc()
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	for serviceAndUri[0:1] == "/" {
		serviceAndUri = serviceAndUri[1:]
	}
	log.Printf("%+v", serviceAndUri)

	// Field 0 is the name of the service
	// The remaining fields, if any, are the URI / query we want to send
	fields := strings.Split(serviceAndUri, "/")
	if len(fields) == 0 {
		brokerRequests.WithLabelValues(
			c.Request.Header.Get(data.HEADER_X_MIGALOO_TAG),
			"",
			"",
			c.Request.Method,
		).Inc()
		brokerResponses.WithLabelValues(
			c.Request.Header.Get(data.HEADER_X_MIGALOO_TAG),
			"",
			"",
			c.Request.Method,
			fmt.Sprint(http.StatusNotFound),
		).Inc()
		c.AbortWithError(http.StatusNotFound, fmt.Errorf("%s: unknown service", c.Request.URL.String()))
		return
	}

	// Get the service
	backend := broker.GetServiceMap().GetBackend(fields[0])
	if backend == nil {
		brokerResponses.WithLabelValues(
			c.Request.Header.Get(data.HEADER_X_MIGALOO_TAG),
			"",
			backend.Label,
			c.Request.Method,
			fmt.Sprint(http.StatusNotFound),
		).Inc()
		c.AbortWithError(http.StatusNotFound, fmt.Errorf("%s: unknown service", c.Request.URL.String()))
		return
	}

	// Update stats
	brokerRequests.WithLabelValues(
		c.Request.Header.Get(data.HEADER_X_MIGALOO_TAG),
		fields[0],
		backend.Label,
		c.Request.Method,
	).Inc()
	brokerCost.WithLabelValues(
		c.Request.Header.Get(data.HEADER_X_MIGALOO_TAG),
		fields[0],
		backend.Label,
		c.Request.Header.Get(data.HEADER_X_MIGALOO_API_CUSTOMER),
	).Add(backend.Cost)

	nowMillis := time.Now().UTC().UnixMilli()
	defer func() {
		brokerRequestsLatency.WithLabelValues(
			c.Request.Header.Get(data.HEADER_X_MIGALOO_TAG),
			fields[0],
			backend.Label,
			c.Request.Method,
		).Add(float64(time.Now().UTC().UnixMilli() - nowMillis))
	}()

	request := *c.Request
	request.URL = backend.Url
	request.Host = backend.Url.Host

	//http: Request.RequestURI can't be set in client requests.
	//http://golang.org/src/pkg/net/http/client.go
	request.RequestURI = ""

	// Make the request
	//
	// TODO(john): put it through the attenuator / circuit breaker etc
	client := http.Client{}
	resp, err := client.Do(&request)
	if err != nil {
		log.Printf("%s: %s: %s", fields[0], backend, err.Error())
		c.Writer.Header().Add(data.HEADER_X_ATTENUATOR_ERROR, err.Error())
		brokerResponses.WithLabelValues(
			c.Request.Header.Get(data.HEADER_X_MIGALOO_TAG),
			"",
			backend.Label,
			c.Request.Method,
			fmt.Sprint(http.StatusBadGateway),
		).Inc()
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}

	defer resp.Body.Close()

	// Send the status
	brokerResponses.WithLabelValues(
		c.Request.Header.Get(data.HEADER_X_MIGALOO_TAG),
		fields[0],
		backend.Label,
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
