package broker

import (
	"fmt"
	"log"
	"math"
	"net/url"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Backend struct {
	Service string   `json:"service"`
	Impl    string   `json:"impl"`
	Label   string   `json:"label"`
	Url     *url.URL `json:"url"`

	// Any parameter mappings
	Params map[string]string `json:"params"`

	// Any headers we want to set in outgoing requests
	Headers map[string]string `json:"params"`

	// Temperature is kind of analagous to load.
	//
	// Backends with a lower temperature are responding more
	// quickly, less loaded etc, have fewer requests in-flight
	// etc.
	//
	// This is to enable smart routing
	Temperature float64 `json:"temperature"`

	// This is for recording requests and responses
	RecordRequestRoot  string `json:"record_request_root"`
	RecordResponseRoot string `json:"record_response_root"`

	windowSizeMillis int64
	stats            chan (*statistic)
}

var backendLatency = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "backend_latency",
		Help:      "The latency for each backend",
	},
	[]string{"service", "label"},
)
var backendStats = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "migaloo",
		Name:      "backend_stats",
		Help:      "Various stats for each backend",
	},
	[]string{"service", "label", "stat"},
)

// NewBackendFromConfig parses a backend from the config.yml
//
// config:
// broker:
//
//	upstream:
//
// # Name of this service.
// # We use this to map /api/v1/broker/{SERVICE} to the service map

// "search":
//
//			  bing:								<-- you are here
//				impl: proxy
//				# The URL to forward-proxy the request to
//				url: https://www.bing.com
//				# Any URL (or POST) params that we want to map
//				params:
//				  # Keep the 'q=' parameter exactly as it is
//				  q: q
//				# Any headers we want to send to the upstream
//				headers:
//				  X-Alizee: wiggle
//		        record:
//	              requests: search
//		          responses: search
//			  google:
//				impl: proxy
//				url: https://www.google.com?q=${q}
//			  rule: random
func NewBackendFromConfig(service string, label string, configMap map[string]interface{}) (backend *Backend, err error) {
	// Which implementation do we use?
	implementation := fmt.Sprint(configMap["impl"])
	switch implementation {
	case "proxy":
		url, err := url.Parse(fmt.Sprint(configMap["url"]))
		if err != nil {
			err = fmt.Errorf("%s: %s: %s: %s", service, label, fmt.Sprint(configMap["url"]), err.Error())
			return backend, err
		}
		backend = &Backend{
			Service: service,
			Label:   label,
			Impl:    implementation,
			Url:     url,
			Params:  make(map[string]string),
			Headers: make(map[string]string),
		}

	default:
		panic(fmt.Sprintf("%s: unknown implementation: '%s'", label, implementation))
	}

	// Deal with any parameter mappings
	if values, hasParams := configMap["params"]; hasParams {
		switch values.(type) {
		case map[string]interface{}:
			for k, v := range values.(map[string]interface{}) {
				backend.Params[k] = fmt.Sprint(v)
			}
		}
	}

	// Any headers we want to set in outgoing requests
	if values, hasHeaders := configMap["headers"]; hasHeaders {
		switch values.(type) {
		case map[string]interface{}:
			for k, v := range values.(map[string]interface{}) {
				backend.Headers[k] = fmt.Sprint(v)
			}
		}
	}

	// If we are recording requests and/or responses, deal with that
	if values, isRecording := configMap["record"]; isRecording {
		switch values.(type) {
		case map[string]interface{}:
			backend.RecordRequestRoot = fmt.Sprint(values.(map[string]interface{})["requests"])
			backend.RecordResponseRoot = fmt.Sprint(values.(map[string]interface{})["responses"])
		}
	}

	return backend, err
}

func NewBackend(service string, label string, url *url.URL) *Backend {
	b := &Backend{
		Service: service,
		Label:   label,
		Url:     url,
		Params:  make(map[string]string),
		Headers: make(map[string]string),

		windowSizeMillis: 1000,
		stats:            make(chan *statistic, 100),
	}
	go b.statsWorker()
	return b
}

type statistic struct {
	when          time.Time
	latencyMillis float64
}

func (b *Backend) String() string {
	return fmt.Sprintf("%s@%s", b.Label, b.Url.String())
}

func (b *Backend) UpdateStats(latencyMillis int64) {
	stat := &statistic{
		when:          time.Now().UTC(),
		latencyMillis: float64(latencyMillis),
	}
	b.stats <- stat
}

func (b *Backend) statsWorker() {
	log.Printf("%s: starting stats worker", b.Label)
	defer log.Printf("%s: terminating stats worker", b.Label)

	previousTime := time.Now().UTC().UnixMilli()
	for {
		stat := <-b.stats
		if stat == nil {
			return
		}

		backendLatency.WithLabelValues(b.Service, b.Label).Add(stat.latencyMillis)
		// Calculate weighted moving average
		timeDifferenceMillis := float64(stat.when.Add(time.Duration(-previousTime)).UnixMilli())
		if timeDifferenceMillis > 0 {
			rate := stat.latencyMillis / timeDifferenceMillis
			weight := 1 - math.Exp(-timeDifferenceMillis/float64(b.windowSizeMillis))
			b.Temperature = weight*rate + (1-weight)*b.Temperature
			backendStats.WithLabelValues(b.Service, b.Label, "temperature").Set(b.Temperature)
		}
		previousTime = stat.when.UnixMilli()
	}
}