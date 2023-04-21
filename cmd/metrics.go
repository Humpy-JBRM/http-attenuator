package cmd

import (
	"http-attenuator/data"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Runs the metrics server",
	Run:   RunMetrics,
}

var metricsAddress string

var proxyRequests *prometheus.CounterVec
var proxyResponses *prometheus.CounterVec
var foo prometheus.Counter

func init() {
	metricsCmd.PersistentFlags().StringVarP(&metricsAddress, "listen", "l", "0.0.0.0:8080", "API listen address (default is 0.0.0.0:8080)")

	proxyRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "migaloo",
			Name:      "proxy_requests",
			Help:      "The number of proxy requests, keyed by host",
		},
		[]string{"host"},
	)

	proxyResponses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "migaloo",
			Name:      "proxy_responses",
			Help:      "The number of proxy responses, keyed by response code and host",
		},
		[]string{"host", "code"},
	)

	foo = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "migaloo",
			Name:      "foo",
			Help:      "The number of proxy responses, keyed by response code and host",
		},
	)
}

func RunMetrics(cmd *cobra.Command, args []string) {
	if viper.GetString(data.CONF_METRICS_LISTEN) != "" {
		metricsAddress = viper.GetString(data.CONF_METRICS_LISTEN)
	}

	log.Printf("Running metrics serve on %s", metricsAddress)
	// Enable metrics
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(metricsAddress, nil)
	if err != nil {
		log.Fatalf("FATAL|cmd.RunMetrics()|Could not start server|%s", err.Error())
	}
	os.Exit(1)
}
