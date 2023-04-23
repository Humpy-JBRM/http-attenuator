package cmd

import (
	"http-attenuator/api"
	gateway "http-attenuator/api/v1/gateway"
	"http-attenuator/data"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var gatewayCmd = &cobra.Command{
	Use:   "gateway",
	Short: "Runs the API gateway",
	Run:   RunGateway,
}

var gatewayAddress string

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
	gatewayEndpoints(ginRouter)

	err = ginRouter.Run(gatewayAddress)
	if err != nil {
		log.Printf("FATAL|cmd.runServer()|Could not start server|%s", err.Error())
	}
}

func gatewayEndpoints(ginRouter *gin.Engine) {
	ginRouter.GET("/api/v1/gateway/*hostAndQuery", gateway.GatewayHandler)
	ginRouter.DELETE("/api/v1/gateway/*hostAndQuery", gateway.GatewayHandler)
	ginRouter.OPTIONS("/api/v1/gateway/*hostAndQuery", gateway.GatewayHandler)
	ginRouter.POST("/api/v1/gateway/*hostAndQuery", gateway.GatewayHandler)
	ginRouter.PUT("/api/v1/gateway/*hostAndQuery", gateway.GatewayHandler)
}
