package data

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var pathologyRequests = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "pathology_requests",
		Help:      "The requests handled by the various pathologies, keyed by name, handler and method",
	},
	[]string{"pathology", "handler", "method"},
)
var pathologyErrors = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "pathology_errors",
		Help:      "The requests handled by the various pathologies, keyed by name, handler and method",
	},
	[]string{"pathology", "handler", "method"},
)
var pathologyLatency = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "pathology_latency",
		Help:      "The latency of the various pathologies, keyed by name, handler and method",
	},
	[]string{"pathology", "handler", "method"},
)
var pathologyResponses = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "pathology_responses",
		Help:      "The responses from the various pathologies, keyed by name, method and status code",
	},
	[]string{"pathology", "handler", "method", "code"},
)
