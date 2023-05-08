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

type EVT interface {
	GetName() string
	GetType() string
	Run() EVTResult
}

type EVTImpl struct {
	evtType string
}

func (e *EVTImpl) GetType() string {
	return e.evtType
}

func (e *EVTImpl) Run() EVTResult {
	return nil
}

type EVTResult interface {
	Error() error

	// How long the test took in millis
	GetDuration() int64
}

func NewEVTResult(err error, durationMillis int64) EVTResult {
	return &EVTResultImpl{
		err:            err,
		durationMillis: durationMillis,
	}
}

type EVTResultImpl struct {
	err            error
	durationMillis int64
}

func (er *EVTResultImpl) Error() error {
	return er.err
}

func (er *EVTResultImpl) GetDuration() int64 {
	return er.durationMillis
}
