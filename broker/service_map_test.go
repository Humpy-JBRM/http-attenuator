package broker

import (
	"os"
	"testing"
)

func TestSimpleGetSync(t *testing.T) {
	os.Setenv("CONFIG_FILE", "../config.yml")
	GetServiceMap()
}
