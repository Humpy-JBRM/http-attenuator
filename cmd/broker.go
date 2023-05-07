package cmd

import (
	"http-attenuator/api"
	broker_api "http-attenuator/api/v1/broker"
	"http-attenuator/broker"
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

	// Register the service broker so it can be picked up by the API
	// handler
	broker.RegisterServiceBroker(appConfig.Config.Broker)

	// All upstream handlers use the service broker
	//
	// TODO(john): there's too much heavy-lifting going on to avoid
	// import loops.  This needs to be fixed.
	for _, upstreamService := range appConfig.Config.Broker.UpstreamFromConfig {
		upstreamService.HandlerFunc = broker.GetServiceBroker().Handle
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
	ginRouter.GET("/api/v1/broker/*serviceAndUri", broker_api.BrokerHandler)
	ginRouter.DELETE("/api/v1/broker/*serviceAndUri", broker_api.BrokerHandler)
	ginRouter.OPTIONS("/api/v1/broker/*serviceAndUri", broker_api.BrokerHandler)
	ginRouter.POST("/api/v1/broker/*serviceAndUri", broker_api.BrokerHandler)
	ginRouter.PUT("/api/v1/broker/*serviceAndUri", broker_api.BrokerHandler)
}
