package api

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var serverRequests = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "server_requests",
		Help:      "The number of server requests, keyed by host and method",
	},
	[]string{"host", "method"},
)
var serverRequestsLatency = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "server_requests_latency",
		Help:      "The latency of server requests, keyed by host and method",
	},
	[]string{"host", "method"},
)
var serverResponses = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "server_responses",
		Help:      "The number of server responses, keyed by host, method and response code",
	},
	[]string{"host", "method"},
)

func ServerHandler(c *gin.Context) {
	panic("IMPLEMENT ME!!!")
}
