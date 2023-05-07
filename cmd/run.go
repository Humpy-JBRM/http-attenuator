package cmd

import (
	"http-attenuator/api"
	"http-attenuator/broker"
	"http-attenuator/data"
	"http-attenuator/server"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs the attenuator",
	Run:   RunRun,
}

var runAddress string

func init() {
	runCmd.PersistentFlags().StringVarP(&runAddress, "listen", "l", "0.0.0.0:8888", "API listen address (default is 0.0.0.0:8888)")
}

func RunRun(cmd *cobra.Command, args []string) {
	if viper.GetString(data.CONF_ATTENUATOR_LISTEN) != "" {
		runAddress = viper.GetString(data.CONF_ATTENUATOR_LISTEN)
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
		log.Fatalf("FATAL|cmd.runServer()|Could not start attenuator|%s", err.Error())
	}

	// Add the run endpoint
	configEndpoints(ginRouter)
	brokerEndpoints(ginRouter)
	gatewayEndpoints(ginRouter)
	serverEndpoints(ginRouter, serverInstance)

	if viper.GetBool(data.CONF_PROXY_ENABLE) {
		go runProxy()
	}

	err = ginRouter.Run(runAddress)
	if err != nil {
		log.Printf("FATAL|cmd.runServer()|Could not start attenuator|%s", err.Error())
	}
}
