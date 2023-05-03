package middleware

import (
	"fmt"
	"http-attenuator/data"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var hsakRequests = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "hsak_requests",
		Help:      "The number of requests to HSAK, keyed by customer API key",
	},
	[]string{"tag", "customer", "endpoint", "method", "code"},
)

// This is a contrived bunch of data, used only for the purposes
// of building out the grafana dashboard and the various billing
// mechanisms
//
// TODO(john): this must be a proper customer database
var customerbyApiKey = map[string]string{
	"":         "(unknown)",
	"69696969": "Ana de Armas",
	"88888888": "Alizee",
	"666":      "GCHQ",
}

// BillingMiddleware is a stub.
//
// We simply disallow anybody who doesnt have a X-Faultmonkey-Api-Key
func BillingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path != "/metrics" {
			customer := customerbyApiKey[c.Request.Header.Get(data.HEADER_X_FAULTMONKEY_API_KEY)]
			c.Request.Header.Add(data.HEADER_X_FAULTMONKEY_API_CUSTOMER, customer)
			hsakRequests.WithLabelValues(
				c.Request.Header.Get(data.HEADER_X_FAULTMONKEY_TAG),
				customer,
				c.Request.URL.Path,
				c.Request.Method,
				"",
			).Inc()
			if c.Request.Header.Get(data.HEADER_X_FAULTMONKEY_API_KEY) == "" {
				c.AbortWithError(
					http.StatusPaymentRequired,
					fmt.Errorf("Invalid API key"),
				)
				return
			}
		}
	}
}
