package cmd

import (
	"http-attenuator/api"
	config "http-attenuator/api/v1/config"
	"http-attenuator/data"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Runs the config API",
	Run:   RunConfig,
}

var configAddress string

func init() {
	configCmd.PersistentFlags().StringVarP(&configAddress, "listen", "l", "0.0.0.0:8888", "API listen address (default is 0.0.0.0:8888)")
}

func RunConfig(cmd *cobra.Command, args []string) {
	if viper.GetString(data.CONF_CONFIG_LISTEN) != "" {
		configAddress = viper.GetString(data.CONF_CONFIG_LISTEN)
	}

	ginRouter, err := api.NewRouter()
	if err != nil {
		log.Fatalf("FATAL|cmd.runServer()|Could not start config API|%s", err.Error())
	}

	// Add the config endpoint
	configEndpoints(ginRouter)

	if viper.GetBool(data.CONF_PROXY_ENABLE) {
		go runProxy()
	}

	err = ginRouter.Run(configAddress)
	if err != nil {
		log.Printf("FATAL|cmd.runServer()|Could not start config API|%s", err.Error())
	}
}

func configEndpoints(ginRouter *gin.Engine) {
	ginRouter.PUT("/api/v1/config/:name/:value", config.SetConfigHandler)
}
