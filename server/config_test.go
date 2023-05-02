package server

import (
	"os"
	"testing"

	"http-attenuator/data"
	config "http-attenuator/facade/config"
)

func TestLoadConfig(t *testing.T) {
	os.Setenv("CONFIG_FILE", "../test_resources/server/pathology_config.yml")
	config.Config()

	appConfig, err := data.LoadConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		t.Fatal(err)
	}
	sb, err := NewServerBuilder().FromConfig(appConfig)
	if err != nil {
		t.Fatal(err)
	}
	server, err := sb.Build()
	if err != nil {
		t.Fatal(err)
	}

	// Validate the server config
	if len(server.Hosts) != 2 {
		t.Errorf("Expected 2 hosts, but got %d", len(server.Hosts))
	}
}
