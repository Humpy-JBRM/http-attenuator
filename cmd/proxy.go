package cmd

import (
	"http-attenuator/api"
	"http-attenuator/data"
	"log"
	"net/http"
	"os"

	"github.com/elazarl/goproxy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Runs the API proxy server",
	Run:   RunProxy,
}

var apiAddress string
var proxyAddress string

func init() {
	proxyCmd.PersistentFlags().StringVarP(&apiAddress, "api", "a", "0.0.0.0:8888", "API listen address (default is 0.0.0.0:8888)")
	proxyCmd.PersistentFlags().StringVarP(&proxyAddress, "proxy", "p", "0.0.0.0:8080", "Proxy listen address (default is 0.0.0.0:8080)")
}

func RunProxy(cmd *cobra.Command, args []string) {
	if viper.GetString(data.CONF_GATEWAY_LISTEN) != "" {
		apiAddress = viper.GetString(data.CONF_GATEWAY_LISTEN)
	}

	ginRouter, err := api.NewRouter()
	if err != nil {
		log.Fatalf("FATAL|cmd.runServer()|Could not start proxy|%s", err.Error())
	}

	// Run gin so we get all of the API endpoints (like /config and /metrics)
	go func() {
		err = ginRouter.Run(apiAddress)
		if err != nil {
			log.Fatalf("FATAL|cmd.runServer()|Could not start proxy|%s", err.Error())
			os.Exit(1)
		}
	}()

	// Now start the proxy
	runProxy()
}

func runProxy() {
	if !viper.GetBool(data.CONF_PROXY_ENABLE) {
		log.Printf("HTTP/s proxy is not enabled ('%s' = false)", data.CONF_PROXY_ENABLE)
		return
	}

	if viper.GetString(data.CONF_PROXY_LISTEN) != "" {
		proxyAddress = viper.GetString(data.CONF_PROXY_LISTEN)
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	log.Printf("Starting HTTP/s proxy on %s", proxyAddress)
	err := http.ListenAndServe(proxyAddress, proxy)
	if err != nil {
		log.Fatalf("FATAL|cmd.runServer()|Could not start proxy|%s", err.Error())
		os.Exit(1)
	}
}
