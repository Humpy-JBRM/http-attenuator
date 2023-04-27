package server

import (
	config "http-attenuator/facade/config"
	"os"
	"reflect"
	"testing"
)

func TestNewPathologyFromConfig(t *testing.T) {
	// Load the config file
	os.Setenv("CONFIG_FILE", "../test_resources/server/pathology_config.yml")
	config.Config()

	err := NewPathologyRegistryFromConfig()
	if err != nil {
		t.Fatal(err)
	}
	if GetPathologyRegistry().GetPathology("simple") == nil {
		t.Fatal("Expected a pathology of 'simple' but got nil")
	}

	// get the 'simple' pathology and check that it is configured
	// exctly as we expect it to be
	simple := GetPathologyRegistry().GetPathology("simple")
	pathologyCdf := simple.(*PathologyImpl).failureModes.failureModes
	if len(pathologyCdf) != 2 {
		t.Fatalf("Expected 2 pathologies in the distribution, but got %d", len(pathologyCdf))
	}

	expectedName := "httpcode"
	actualName := pathologyCdf[0].(*FailureModeImpl).name
	if expectedName != actualName {
		t.Errorf("Expected name='%s', but got '%s'", expectedName, actualName)
	}
	expectedWeight := int64(90)
	actualWeight := pathologyCdf[0].(*FailureModeImpl).weight
	if expectedWeight != actualWeight {
		t.Errorf("Expected weight='%d', but got '%d'", expectedWeight, actualWeight)
	}
	expectedCdf := float64(0.9)
	actualCdf := pathologyCdf[0].(*FailureModeImpl).cdf
	if expectedCdf != actualCdf {
		t.Errorf("Expected cdf='%f', but got '%f'", expectedCdf, actualCdf)
	}

	// We should have a httpcode FailureMode, with a handler
	var httpCodeFailureMode *FailureModeImpl
	for i := 0; i < len(pathologyCdf); i++ {
		if pathologyCdf[i].(*FailureModeImpl).name == "httpcode" {
			httpCodeFailureMode = pathologyCdf[i].(*FailureModeImpl)
			break
		}
	}
	if httpCodeFailureMode == nil {
		t.Fatal("Expected to find a httpcode failure mode, but did not")
	}

	// This failure mode should have a HttpCodeHandler
	if httpCodeFailureMode.handler == nil {
		t.Fatal("httpcode failure mode has no handler")
	}
	if _, isHttpCodeHandler := httpCodeFailureMode.handler.(*HttpCodeHandler); !isHttpCodeHandler {
		t.Fatalf("httpcode failure mode has handler %s, but expected HttpCodeHandler", reflect.TypeOf(httpCodeFailureMode.handler))
	}
	handler := httpCodeFailureMode.handler.(*HttpCodeHandler)

	// This handler should have 5 entries in the CDF
	if len(handler.cdf) != 5 {
		t.Errorf("Expected httpcode hander to have 5 cdf entries but it has %d", len(handler.cdf))
	}

	// The first one should be for the http 200 code
	if handler.cdf[0].(*HttpCodeCdf).code != 200 {
		t.Errorf("Expected first cdf entry to be for code 200, but it is %d", handler.cdf[0].(*HttpCodeCdf).code)
	}
	http200CdfEntry := handler.cdf[0].(*HttpCodeCdf)
	if http200CdfEntry.cdf != 0.8 {
		t.Errorf("Expected first cdf entry to be for code 200 with cdf = 0.8, but it is %d with cdf = %f", handler.cdf[0].(*HttpCodeCdf).code, http200CdfEntry.cdf)
	}
	for _, c := range handler.cdf {
		t.Errorf("%+v", c)
	}

	expectedName = "timeout"
	actualName = pathologyCdf[1].(*FailureModeImpl).name
	if expectedName != actualName {
		t.Errorf("Expected name='%s', but got '%s'", expectedName, actualName)
	}
	expectedWeight = int64(10)
	actualWeight = pathologyCdf[1].(*FailureModeImpl).weight
	if expectedWeight != actualWeight {
		t.Errorf("Expected weight='%d', but got '%d'", expectedWeight, actualWeight)
	}
	expectedCdf = float64(1.0)
	actualCdf = pathologyCdf[1].(*FailureModeImpl).cdf
	if expectedCdf != actualCdf {
		t.Errorf("Expected cdf='%f', but got '%f'", expectedCdf, actualCdf)
	}

	choices := make(map[string]int)
	for count := 100; count > 0; count-- {
		failureMode := simple.ChooseFailureMode()
		if failureMode == nil {
			t.Fatalf("%s.ChooseFailureMode():Got nil failure mode", simple.(*PathologyImpl).name)
		}
		choices[failureMode.(*FailureModeImpl).name]++
	}

	t.Fatal("TODO(john): timeout configuration")

	// Out of 100 choices, we should have ~90% 'httpcode' and ~10% timeout
	upper := 95
	lower := 85
	choice := "httpcode"
	if choices[choice] < lower || choices[choice] > 95 {
		t.Errorf("Expected %d <= %s <= %d, but got %d", lower, choice, upper, choices[choice])
	}
	upper = 15
	lower = 5
	choice = "timeout"
	if choices[choice] < lower || choices[choice] > upper {
		t.Errorf("Expected %d <= %s <= %d, but got %d", lower, choice, upper, choices[choice])
	}
}
