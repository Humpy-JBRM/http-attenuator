package cmd

import (
	"fmt"
	"http-attenuator/data"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "hsak",
	Short: "HSAK (HTTP Swiss Army Knife) is a smart API gateway and proxy",
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

var cfgFile string
var appConfig *data.AppConfig

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "f", "config.yml", "config file (default is config.yml)")

	rootCmd.AddCommand(brokerCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(gatewayCmd)
	rootCmd.AddCommand(proxyCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(serverCmd)

	// Microservices
}

func initConfig() {
	if configFile, isSet := os.LookupEnv("CONFIG_FILE"); isSet {
		cfgFile = configFile
	}
	if cfgFile != "" {
		err := LoadConfig(cfgFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	// util.StartPrometheus()
}

func LoadConfig(cfgFile string) error {
	log.Printf("INFO|cmd.initConfig()|Reading config from %s|", cfgFile)
	cfDir := filepath.Dir(cfgFile)
	if cfDir == "" {
		cfDir = "."
	}
	viper.AddConfigPath(cfDir)
	viper.SetConfigType(filepath.Ext(cfgFile)[1:])
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("ERROR|cmd.LoadConfig()|Could not read file %s|%s", viper.ConfigFileUsed(), err.Error())
	}

	// Load the structured config
	var err error
	appConfig, err = data.LoadConfig(viper.ConfigFileUsed())
	if err != nil {
		return fmt.Errorf("cmd.LoadConfig(%s): %s", viper.ConfigFileUsed(), err.Error())
	}

	for _, key := range viper.AllKeys() {
		log.Printf("%s: %v\n", key, viper.Get(key))
	}

	return nil
}
