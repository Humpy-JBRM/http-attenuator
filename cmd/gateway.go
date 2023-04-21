package cmd

import (
	"encoding/json"
	"http-attenuator/api"
	"http-attenuator/data"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var gatewayCmd = &cobra.Command{
	Use:   "gateway",
	Short: "Runs the API gateway",
	Run:   RunGateway,
}

var gatewayAddress string

var gatewayRequests = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "gateway_requests",
		Help:      "The number of gateway requests, keyed by host",
	},
	[]string{"host"},
)

var gatewayResponses = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "gateway_responses",
		Help:      "The number of gateway responses, keyed by response code and host",
	},
	[]string{"host", "code"},
)

func init() {
	gatewayCmd.PersistentFlags().StringVarP(&gatewayAddress, "listen", "l", "0.0.0.0:8888", "API listen address (default is 0.0.0.0:8002)")
}

func RunGateway(cmd *cobra.Command, args []string) {
	if viper.GetString(data.CONF_GATEWAY_LISTEN) != "" {
		gatewayAddress = viper.GetString(data.CONF_GATEWAY_LISTEN)
	}

	ginRouter, err := api.NewRouter()
	if err != nil {
		log.Fatalf("FATAL|cmd.runServer()|Could not start server|%s", err.Error())
	}

	// Add the gateway endpoint
	ginRouter.GET("/api/v1/gateway/*hostAndQuery", gatewayHandler)
	ginRouter.DELETE("/api/v1/gateway/*hostAndQuery", gatewayHandler)
	ginRouter.OPTIONS("/api/v1/gateway/*hostAndQuery", gatewayHandler)
	ginRouter.POST("/api/v1/gateway/*hostAndQuery", gatewayHandler)
	ginRouter.PUT("/api/v1/gateway/*hostAndQuery", gatewayHandler)

	err = ginRouter.Run(gatewayAddress)
	if err != nil {
		log.Printf("FATAL|cmd.runServer()|Could not start server|%s", err.Error())
	}
}

func gatewayHandler(c *gin.Context) {
	// Extract the host from the URL
	hostAndQuery := c.Param("hostAndQuery")
	if hostAndQuery == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	for hostAndQuery[0:1] == "/" {
		hostAndQuery = hostAndQuery[1:]
	}
	log.Printf("%+v", hostAndQuery)

	hostAndQueryUrl, err := url.Parse(hostAndQuery)
	if err != nil {
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
	rb, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		log.Println(err.Error())
	}
	log.Println(string(rb))
	resp, err := client.Do(&request)
	if err != nil {
		log.Printf("%s: %s", hostAndQuery, err.Error())
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}

	defer resp.Body.Close()

	// Send the status
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
