package broker

import (
	config "http-attenuator/facade/config"
	"os"
	"testing"
)

func TestBackendForwardProxy(t *testing.T) {
	os.Setenv("CONFIG_FILE", "../test_resources/broker/config.yml")
	config.Config()
}
