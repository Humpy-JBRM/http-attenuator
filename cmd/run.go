package cmd

import (
	"http-attenuator/api"
	"http-attenuator/data"
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

	ginRouter, err := api.NewRouter()
	if err != nil {
		log.Fatalf("FATAL|cmd.runServer()|Could not start attenuator|%s", err.Error())
	}

	// Add the run endpoint
	configEndpoints(ginRouter)
	brokerEndpoints(ginRouter)
	gatewayEndpoints(ginRouter)

	if viper.GetBool(data.CONF_PROXY_ENABLE) {
		go runProxy()
	}

	err = ginRouter.Run(runAddress)
	if err != nil {
		log.Printf("FATAL|cmd.runServer()|Could not start attenuator|%s", err.Error())
	}
}
