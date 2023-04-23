package cmd

import (
	"http-attenuator/api"
	broker "http-attenuator/api/v1/broker"
	"http-attenuator/data"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var brokerCmd = &cobra.Command{
	Use:   "broker",
	Short: "Runs the service broker",
	Run:   RunBroker,
}

var brokerAddress string

func init() {
	brokerCmd.PersistentFlags().StringVarP(&brokerAddress, "listen", "l", "0.0.0.0:8888", "API listen address (default is 0.0.0.0:8888)")
}

func RunBroker(cmd *cobra.Command, args []string) {
	if viper.GetString(data.CONF_BROKER_LISTEN) != "" {
		brokerAddress = viper.GetString(data.CONF_BROKER_LISTEN)
	}

	ginRouter, err := api.NewRouter()
	if err != nil {
		log.Fatalf("FATAL|cmd.runServer()|Could not start service broker|%s", err.Error())
	}

	// Add the endpoints
	configEndpoints(ginRouter)

	// Add the broker endpoints
	brokerEndpoints(ginRouter)

	if viper.GetBool(data.CONF_PROXY_ENABLE) {
		go runProxy()
	}

	err = ginRouter.Run(brokerAddress)
	if err != nil {
		log.Printf("FATAL|cmd.runServer()|Could not start service broker|%s", err.Error())
	}
}

func brokerEndpoints(ginRouter *gin.Engine) {
	ginRouter.GET("/api/v1/broker/*serviceAndUri", broker.BrokerHandler)
	ginRouter.DELETE("/api/v1/broker/*serviceAndUri", broker.BrokerHandler)
	ginRouter.OPTIONS("/api/v1/broker/*serviceAndUri", broker.BrokerHandler)
	ginRouter.POST("/api/v1/broker/*serviceAndUri", broker.BrokerHandler)
	ginRouter.PUT("/api/v1/broker/*serviceAndUri", broker.BrokerHandler)
}
