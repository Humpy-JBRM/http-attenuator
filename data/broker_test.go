package data

import (
	config "http-attenuator/facade/config"
	"os"
	"testing"
)

func TestBrokerFromConfig(t *testing.T) {
	os.Setenv("CONFIG_FILE", "../test_resources/broker/config.yml")
	config.Config()

	appConfig, err := LoadConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		t.Fatal(err)
	}

	broker := appConfig.Config.Broker
	expectedListen := "0.0.0.0:8888"
	actualListen := broker.Listen
	if expectedListen != actualListen {
		t.Errorf("Expected listen='%s', but got '%s'", expectedListen, actualListen)
	}
	expectedLen := 3
	actualLen := len(broker.upstream)
	if expectedLen != actualLen {
		t.Errorf("Expected %d upstream services, but got %d", expectedLen, actualLen)
	}

	t.Fatal("STILL WORKING ON BROKER CONFIG")
}
