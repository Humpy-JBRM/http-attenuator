package broker

import (
	"fmt"
	"http-attenuator/data"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type ServiceBroker interface {
	Handle(c *gin.Context)
}

type ServiceBrokerImpl struct {
	data.BrokerImpl

	// TODO(john): upstream temperature

	// TODO(john): rate-limiting / rate detection
}

var brokerRequests = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "broker_requests",
		Help:      "The number of broker requests, keyed by upstream/backend",
	},
	[]string{"tag", "upstream", "backend", "method"},
)
var brokerErrors = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "broker_errors",
		Help:      "The number of broker errors, keyed by upstream/backend",
	},
	[]string{"tag", "upstream", "backend", "method"},
)
var brokerRequestsLatency = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "broker_requests_latency",
		Help:      "The latency of broker requests, keyed by upstream/backend",
	},
	[]string{"tag", "upstream", "backend", "method"},
)
var brokerResponses = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "broker_responses",
		Help:      "The number of broker responses, keyed by response code and upstream/backend",
	},
	[]string{"tag", "upstream", "backend", "method", "code"},
)

var serviceBrokerInstance ServiceBroker
var serviceBrokerOnce sync.Once

func RegisterServiceBroker(broker data.Broker) {
	serviceBrokerOnce.Do(func() {
		serviceBrokerInstance = &ServiceBrokerImpl{
			BrokerImpl: *broker.(*data.BrokerImpl),
		}
	})
}

func GetServiceBroker() ServiceBroker {
	return serviceBrokerInstance
}

func (sb *ServiceBrokerImpl) Handle(c *gin.Context) {
	// Extract the service from the URL
	serviceAndUri := c.Param("serviceAndUri")
	if serviceAndUri == "" {
		brokerRequests.WithLabelValues(
			c.Request.Header.Get(data.HEADER_X_FAULTMONKEY_TAG),
			"", // upstream
			"", // backend
			c.Request.Method,
		).Inc()
		brokerResponses.WithLabelValues(
			c.Request.Header.Get(data.HEADER_X_FAULTMONKEY_TAG),
			"", // upstream
			"", // backend
			c.Request.Method,
			fmt.Sprint(http.StatusNotFound),
		).Inc()
		brokerErrors.WithLabelValues(
			c.Request.Header.Get(data.HEADER_X_FAULTMONKEY_TAG),
			"", // upstream
			"", // backend
			c.Request.Method,
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
			c.Request.Header.Get(data.HEADER_X_FAULTMONKEY_TAG),
			"", // upstream
			"", // backend
			c.Request.Method,
		).Inc()
		brokerResponses.WithLabelValues(
			c.Request.Header.Get(data.HEADER_X_FAULTMONKEY_TAG),
			"", // upstream
			"", // backend
			c.Request.Method,
			fmt.Sprint(http.StatusNotFound),
		).Inc()
		brokerErrors.WithLabelValues(
			c.Request.Header.Get(data.HEADER_X_FAULTMONKEY_TAG),
			"", // upstream
			"", // backend
			c.Request.Method,
		).Inc()
		c.AbortWithError(http.StatusNotFound, fmt.Errorf("%s: unknown service", c.Request.URL.String()))
		return
	}

	nowMillis := time.Now().UTC().UnixMilli()
	var upstream data.Upstream
	defer func() {
		brokerRequestsLatency.WithLabelValues(
			c.Request.Header.Get(data.HEADER_X_FAULTMONKEY_TAG),
			upstream.GetName(),
			"", // backend
			c.Request.Method,
		).Add(float64(time.Now().UTC().UnixMilli() - nowMillis))
	}()
	// Get an upstream
	upstream = sb.GetUpstream(fields[0])
	if upstream == nil {
		err := fmt.Errorf("%s.Handle(%s): No upstream for '%s'", sb.GetName(), c.Request.URL, fields[0])
		log.Println(err)
		brokerErrors.WithLabelValues(
			c.Request.Header.Get(data.HEADER_X_FAULTMONKEY_TAG),
			"", // upstream
			"", // backend
			c.Request.Method,
		).Inc()
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	upstream.Handle(c)
}

// NewBroker returns a simple reverse proxy which encapsulates
// a bunch of backend services
func NewBroker(brokerFromConfig data.Broker) data.Broker {
	// TODO(john): Are we recording?
	// recordRequestsRoot, _ := config.Config().GetString(data.CONF_GATEWAY_RECORD_REQUESTS)
	// recordResponsesRoot, _ := config.Config().GetString(data.CONF_GATEWAY_RECORD_REQUESTS)

	serviceBroker := &ServiceBrokerImpl{
		BrokerImpl: *(brokerFromConfig.(*data.BrokerImpl)),
	}

	return serviceBroker
}
