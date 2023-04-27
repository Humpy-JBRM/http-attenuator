package cmd

import (
	"http-attenuator/api"
	"http-attenuator/data"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Runs the server to simulate the various failure modes",
	Run:   RunGateway,
}

var serverAddress string

func init() {
	serverCmd.PersistentFlags().StringVarP(&serverAddress, "listen", "l", "0.0.0.0:8888", "API listen address (default is 0.0.0.0:8888)")
}

func RunServer(cmd *cobra.Command, args []string) {
	if viper.GetString(data.CONF_SERVER_LISTEN) != "" {
		serverAddress = viper.GetString(data.CONF_SERVER_LISTEN)
	}

	ginRouter, err := api.NewRouter()
	if err != nil {
		log.Fatalf("FATAL|cmd.runServer()|Could not start server|%s", err.Error())
	}

	// Add the server endpoint
	serverEndpoints(ginRouter)

	err = ginRouter.Run(serverAddress)
	if err != nil {
		log.Printf("FATAL|cmd.runServer()|Could not start server|%s", err.Error())
	}
}

func serverEndpoints(ginRouter *gin.Engine) {
	ginRouter.NoRoute()
	panic("TODO(john): spin up the naughty server")
}

func loadServerConfig(cmd *cobra.Command, args []string) {

}
