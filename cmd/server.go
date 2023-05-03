package cmd

import (
	"http-attenuator/api"
	"http-attenuator/data"
	config "http-attenuator/facade/config"
	"http-attenuator/server"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Runs the FaultMonkey server to simulate the various failure modes",
	Run:   RunServer,
}

func init() {
}

func RunServer(cmd *cobra.Command, args []string) {
	enabled, err := config.Config().GetBool(data.CONF_SERVER_ENABLE)
	if err != nil {
		log.Fatalf("cmd.runServer(): Could not start server|%s", err.Error())
	}

	// Short-circuit if the server is not enabled
	if !enabled {
		log.Printf("cmd.runServer(): Server is not enabled|%s", err.Error())
		return
	}

	// Configure the server for running.
	//
	// This takes care of any backpatching, config validation etc.
	serverBuilder, err := server.NewServerBuilder().FromConfig(appConfig)
	if err != nil {
		log.Fatalf("cmd.runServer(): %s", err.Error())
	}
	serverInstance, err := serverBuilder.Build()
	if err != nil {
		log.Fatalf("cmd.runServer(): %s", err.Error())
	}

	ginRouter, err := api.NewRouter()
	if err != nil {
		log.Fatalf("FATAL|cmd.runServer()|Could not start server|%s", err.Error())
	}

	// Add the server endpoint
	serverEndpoints(ginRouter, serverInstance)

	err = ginRouter.Run(serverInstance.Listen)
	if err != nil {
		log.Printf("FATAL|cmd.runServer()|Could not start server|%s", err.Error())
	}
}

func serverEndpoints(ginRouter *gin.Engine, serverInstance *data.Server) {
	faultMonkey := server.NewFaultMonkey(serverInstance)
	ginRouter.Use(faultMonkey.Handle)
}
