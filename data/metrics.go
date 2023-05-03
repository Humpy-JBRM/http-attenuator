package data

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var pathologyRequests = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "pathology_requests",
		Help:      "The requests handled by the various pathologies, keyed by name, handler and method",
	},
	[]string{"profile", "pathology", "host", "method", "code"},
)
var pathologyErrors = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "pathology_errors",
		Help:      "The requests handled by the various pathologies, keyed by name, handler and method",
	},
	[]string{"profile", "pathology", "host", "method", "code"},
)
var pathologyLatency = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "pathology_latency",
		Help:      "The latency of the various pathologies, keyed by name, handler and method",
	},
	[]string{"profile", "pathology", "host", "method", "code"},
)
var pathologyResponses = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "pathology_responses",
		Help:      "The responses from the various pathologies, keyed by name, method and status code",
	},
	[]string{"profile", "pathology", "host", "method", "code"},
)
