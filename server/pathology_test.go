package server

import (
	"http-attenuator/data"
	config "http-attenuator/facade/config"
	"os"
	"testing"
)

func TestNewPathologyFromConfig(t *testing.T) {
	// Load the config file
	os.Setenv("CONFIG_FILE", "../test_resources/server/pathology_config.yml")
	config.Config()

	pathologyRoot, err := config.Config().GetAllValues(data.CONF_PATHOLOGY)
	if err != nil {
		t.Fatal(err)
	}
	err = NewPathologyRegistryFromConfig(pathologyRoot)
	if err != nil {
		t.Fatal(err)
	}
	if GetPathologyRegistry().GetPathology("simple") == nil {
		t.Fatal("Expected a pathology of 'simple' but got nil")
	}

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
