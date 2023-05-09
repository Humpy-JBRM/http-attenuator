package evt

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var evtExecuted = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "evt",
		Help:      "EVTs run, keyed by type and name",
	},
	[]string{"type", "name"},
)
var evtDuration = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "evt_duration",
		Help:      "Duration of the EVTs run, keyed by type and name",
	},
	[]string{"type", "name"},
)
var evtErrors = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "evt_failure",
		// TODO(john): error classifier
		Help: "EVT failures, keyed by type and name",
	},
	[]string{"type", "name", "class"}, // 'class' is the class of failure, as per TODO classifier
)
