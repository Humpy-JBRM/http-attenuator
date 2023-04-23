package broker

import (
	"os"
	"testing"
)

func TestLoadUpstreamConfig(t *testing.T) {
	os.Setenv("CONFIG_FILE", "../test_resources/broker/config.yml")
	serviceMap := GetServiceMap()
	if serviceMap == nil {

	}
}
