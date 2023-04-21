package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var gatewayRequests = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "gateway_requests",
		Help:      "The number of gateway requests, keyed by host",
	},
	[]string{"tag", "host", "method"},
)
var gatewayRequestsLatency = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "gateway_requests_latency",
		Help:      "The latency of gateway requests, keyed by host",
	},
	[]string{"tag", "host", "method"},
)

var gatewayResponses = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "gateway_responses",
		Help:      "The number of gateway responses, keyed by response code and host",
	},
	[]string{"tag", "host", "method", "code"},
)

func GatewayHandler(c *gin.Context) {
	nowMillis := time.Now().UnixMilli()
	defer func() {
		gatewayRequestsLatency.WithLabelValues(
			c.Request.Header.Get("X-migaloo-tag"),
			c.Request.Host,
			c.Request.Method,
		).Add(float64(time.Now().UnixMilli() - nowMillis))
	}()

	// Extract the host from the URL
	gatewayRequests.WithLabelValues(
		c.Request.Header.Get("X-migaloo-tag"),
		c.Request.Host,
		c.Request.Method,
	).Inc()
	hostAndQuery := c.Param("hostAndQuery")
	if hostAndQuery == "" {
		gatewayResponses.WithLabelValues(
			c.Request.Header.Get("X-migaloo-tag"),
			c.Request.Host,
			c.Request.Method,
			fmt.Sprint(http.StatusNotFound),
		).Inc()
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	for hostAndQuery[0:1] == "/" {
		hostAndQuery = hostAndQuery[1:]
	}
	log.Printf("%+v", hostAndQuery)

	hostAndQueryUrl, err := url.Parse(hostAndQuery)
	if err != nil {
		gatewayResponses.WithLabelValues(
			c.Request.Header.Get("X-migaloo-tag"),
			c.Request.Host,
			c.Request.Method,
			fmt.Sprint(http.StatusBadRequest),
		).Inc()
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	request := *c.Request
	request.URL = hostAndQueryUrl
	request.Host = hostAndQueryUrl.Host

	//http: Request.RequestURI can't be set in client requests.
	//http://golang.org/src/pkg/net/http/client.go
	request.RequestURI = ""

	// Make the request
	//
	// TODO(john): put it through the attenuator / circuit breaker etc
	client := http.Client{}
	resp, err := client.Do(&request)
	if err != nil {
		log.Printf("%s: %s", hostAndQuery, err.Error())
		gatewayResponses.WithLabelValues(
			c.Request.Header.Get("X-migaloo-tag"),
			c.Request.Host,
			c.Request.Method,
			fmt.Sprint(http.StatusBadGateway),
		).Inc()
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}

	defer resp.Body.Close()

	// Send the status
	gatewayResponses.WithLabelValues(
		c.Request.Header.Get("X-migaloo-tag"),
		c.Request.Host,
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
