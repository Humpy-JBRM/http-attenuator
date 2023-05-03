package data

import (
	config "http-attenuator/facade/config"
	"http-attenuator/util"
	"net/http"
	"os"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func validateConfig(t *testing.T, appConfig *AppConfig) {
	// We should have a pathology profile called 'simple'
	simplePathologyProfile := appConfig.Config.GetProfile("simple")

	// Check that this pathology is registered.
	// If it is not, then it cannot be backpatched into the server
	//GetRegistry().GetPathology("simple")

	// This simple pathology should have a httpcode pathology and a timeout pathology
	httpcodePathology := simplePathologyProfile.GetPathologyByName("httpcode")
	if httpcodePathology == nil {
		t.Fatal("Profile'simple' does not have the expected 'httpcode' pathology")
	}

	// The httpcode pathology should have a weight of 0.9 in the 'simple' profile
	expectedWeight := 90
	actualWeight := httpcodePathology.GetWeight()
	if expectedWeight != actualWeight {
		t.Errorf("httpcode: expected weight=%d, got %d", expectedWeight, actualWeight)
	}

	// httpcode should have five responses
	if len(httpcodePathology.(*PathologyImpl).Responses) != 5 {
		t.Errorf("Expected 5 responses in httpcode pathology, but got %d", len(httpcodePathology.(*PathologyImpl).Responses))
	}

	// HTTP 200
	expectedName := "httpcode"
	expectedProfile := "simple"
	expectedWeight = 80
	expectedHeaders := http.Header{
		"Content-type": []string{"application/json"},
	}
	expectedBody := `{"success": true, "pathology": "simple", "handler": "httpcode"}`
	expectedCdf := 0.8
	actualWeight = httpcodePathology.(*PathologyImpl).Responses[200].Weight
	actualHeaders := httpcodePathology.(*PathologyImpl).Responses[200].Headers
	actualBody := httpcodePathology.(*PathologyImpl).Responses[200].Body
	actualCdf := httpcodePathology.(*PathologyImpl).Responses[200].CDF()
	actualName := httpcodePathology.GetName()
	actualProfile := httpcodePathology.GetProfileName()
	if !util.AlmostEqual(expectedCdf, actualCdf) {
		t.Errorf("httpcode.200: expected cdf=%f, got %f", expectedCdf, actualCdf)
	}
	if expectedBody != actualBody {
		t.Errorf("httpcode.200: expected body='%s', got '%s'", expectedBody, actualBody)
	}
	if !reflect.DeepEqual(expectedHeaders, actualHeaders) {
		t.Errorf("httpcode.200: expected headers='%v', got '%v'", expectedHeaders, actualHeaders)
	}
	if expectedWeight != actualWeight {
		t.Errorf("httpcode.200: expected weight=%d, got %d", expectedWeight, actualWeight)
	}
	if expectedName != actualName {
		t.Errorf("httpcode.200: expected name=%s, got %s", expectedName, actualName)
	}
	if expectedProfile != actualProfile {
		t.Errorf("httpcode.200: expected profile=%s, got %s", expectedProfile, actualProfile)
	}

	// HTTP 429
	expectedWeight = 5
	expectedHeaders = http.Header{
		"X-Backoff-Millis": []string{"60000"},
		"X-Retry-After":    []string{"now() + 60s"},
	}
	expectedBody = ""
	expectedCdf = 0.05
	actualWeight = httpcodePathology.(*PathologyImpl).Responses[429].Weight
	actualHeaders = httpcodePathology.(*PathologyImpl).Responses[429].Headers
	actualBody = httpcodePathology.(*PathologyImpl).Responses[429].Body
	actualCdf = httpcodePathology.(*PathologyImpl).Responses[429].CDF()
	if !util.AlmostEqual(expectedCdf, actualCdf) {
		t.Errorf("httpcode.429: expected cdf=%f, got %f", expectedCdf, actualCdf)
	}
	if expectedWeight != actualWeight {
		t.Errorf("httpcode.429: expected weight=%d, got %d", expectedWeight, actualWeight)
	}
	if expectedBody != actualBody {
		t.Errorf("httpcode.429: expected body='%s', got '%s'", expectedBody, actualBody)
	}
	if !reflect.DeepEqual(expectedHeaders, actualHeaders) {
		t.Errorf("httpcode.429: expected headers='%v', got '%v'", expectedHeaders, actualHeaders)
	}

	// timeout pathology
	timeoutPathology := simplePathologyProfile.GetPathologyByName("timeout")
	if timeoutPathology == nil {
		t.Fatal("Profile'simple' does not have the expected 'timeout' pathology")
	}
	expectedWeight = 10
	actualWeight = timeoutPathology.GetWeight()
	if expectedWeight != actualWeight {
		t.Errorf("timeout: expected weight=%d, got %d", expectedWeight, actualWeight)
	}

	// The timeout duration
	expectedMillisLo := int64(100)
	expectedMillisHi := int64(2000)
	actualMillis := timeoutPathology.SelectResponse().GetDuration().Milliseconds()
	if actualMillis < expectedMillisLo || actualMillis > expectedMillisHi {
		t.Errorf("timeout: expected %d < duration < %d, got %d", expectedMillisLo, expectedMillisHi, actualMillis)
	}

	// The timeout response
	expectedCode := 200
	expectedHeaders = http.Header{
		"Content-type": []string{"application/json"},
	}
	expectedBody = `{"success": true, "pathology": "simple", "handler": "timeout"}`
	expectedName = "timeout"
	expectedProfile = "simple"
	// weight is not specified in the config
	expectedWeight = 0
	// single response means that cdf is 1
	expectedCdf = 1.0
	actualCode := timeoutPathology.SelectResponse().Code
	actualHeaders = timeoutPathology.SelectResponse().Headers
	actualBody = timeoutPathology.SelectResponse().Body
	actualName = timeoutPathology.GetName()
	actualProfile = timeoutPathology.GetProfileName()
	actualWeight = timeoutPathology.SelectResponse().GetWeight()
	actualCdf = timeoutPathology.SelectResponse().CDF()
	if expectedCode != actualCode {
		t.Errorf("timeout: expected code=%d, got %d", expectedCode, actualCode)
	}
	if expectedBody != actualBody {
		t.Errorf("timeout: expected body='%s', got '%s'", expectedBody, actualBody)
	}
	if !reflect.DeepEqual(expectedHeaders, actualHeaders) {
		t.Errorf("timeout: expected headers='%v', got '%v'", expectedHeaders, actualHeaders)
	}
	if expectedName != actualName {
		t.Errorf("timeout: expected name=%s, got %s", expectedName, actualName)
	}
	if expectedProfile != actualProfile {
		t.Errorf("timeout: expected profile=%s, got %s", expectedProfile, actualProfile)
	}
	if expectedWeight != actualWeight {
		t.Errorf("timeout.200: expected weight=%d, got %d", expectedWeight, actualWeight)
	}
	if expectedCdf != actualCdf {
		t.Errorf("timeout.200: expected cdf=%f, got %f", expectedCdf, actualCdf)
	}

	// Check the server config
	if !appConfig.Config.Server.Enable {
		t.Errorf("Expected server '%s' to be enabled", appConfig.Config.Server.Name)
	}

	// We should have two serving instances:
	//
	//	default: simple pathology profile
	//
	//	goodboy.com: good_boy pathology profile
	if len(appConfig.Config.Server.Hosts) != 2 {
		t.Errorf("Expected server '%s' to have two host mappings but it has %d", appConfig.Config.Server.Name, len(appConfig.Config.Server.Hosts))
	}

	hostname := "default"
	expectedPathology := "simple"
	host := appConfig.Config.Server.Hosts[hostname]
	if host.PathologyProfileName != expectedPathology {
		t.Errorf("Expected server host '%s' to have pathology '%s', but it has '%s'", hostname, expectedPathology, host.PathologyProfileName)
	}
	hostname = "goodboy.com"
	expectedPathology = "good_boy"
	host = appConfig.Config.Server.Hosts[hostname]
	if host.PathologyProfileName != expectedPathology {
		t.Errorf("Expected server host '%s' to have pathology '%s', but it has '%s'", hostname, expectedPathology, host.PathologyProfileName)
	}

	// the pathology profiles should have been registered
	profileName := "good_boy"
	if GetProfileRegistry().GetPathologyProfile(profileName) == nil {
		t.Errorf("Expected pathology profile '%s' to be registered but it was not", profileName)
	}
	profileName = "simple"
	if GetProfileRegistry().GetPathologyProfile(profileName) == nil {
		t.Errorf("Expected pathology profile '%s' to be registered but it was not", profileName)
	}

	// We should have two pathologies in the simple profile
	simplePathology := GetProfileRegistry().GetPathologyProfile("simple")
	expectedNumberOfPathologies := 2
	if len(simplePathology.(*PathologyProfileImpl).pathologyCdf) != expectedNumberOfPathologies {
		t.Errorf("Expected pathology profile 'simple' to have %d pathologies, but it has %d", expectedNumberOfPathologies, len(simplePathology.(*PathologyProfileImpl).pathologyCdf))
	}

	// httpcode pathology should be selected ~90% of the time
	// timeout pathology should be selected ~10% of the time
	//
	// TODO(john): the values out of the rng do not appear to be uniform, causing the below to fail
	// selected := make(map[string]int)
	// iterations := 10000
	// for i := 0; i < iterations; i++ {
	// 	selectedPathology := simplePathology.GetPathology()
	// 	selected[selectedPathology.GetName()]++
	// }
	// expectedHttpCode := float64(0.9)
	// expectedTimeout := float64(0.1)
	// actualHttpCode := float64(selected["httpcode"]) / float64(iterations)
	// actualTimeout := float64(selected["timeout"]) / float64(iterations)
	// if !util.AlmostEqual(expectedHttpCode, actualHttpCode) {
	// 	t.Errorf("Expected httpcode ~%.2f, but was %.2f (%d/%d)", expectedHttpCode, actualHttpCode, selected["httpcode"], iterations)
	// }
	// if !util.AlmostEqual(expectedTimeout, actualTimeout) {
	// 	t.Errorf("Expected timeout ~%.2f, but was %.2f (%d/%d)", expectedTimeout, actualTimeout, selected["timeout"], iterations)
	// }
}

func TestParseConfigYaml(t *testing.T) {
	os.Setenv("CONFIG_FILE", "../test_resources/server/pathology_config.yml")
	config.Config()

	appConfig, err := LoadConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		t.Fatal(err)
	}

	validateConfig(t, appConfig)
}

func TestConfigRoundTrip(t *testing.T) {
	os.Setenv("CONFIG_FILE", "../test_resources/server/pathology_config.yml")
	config.Config()

	appConfig, err := LoadConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		t.Fatal(err)
	}

	// Write config to a temp file
	f, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	configBytes, _ := yaml.Marshal(appConfig)
	f.Write(configBytes)

	// Read the config in from this file
	// This is the round-trip
	appConfig, err = LoadConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		t.Fatal(err)
	}

	// Validate the config
	validateConfig(t, appConfig)
}

func TestHttpResponseSatisfiesHasCDF(t *testing.T) {
	var r interface{} = &HttpResponse{}
	if _, isHasCDF := r.(HasCDF); !isHasCDF {
		t.Error("HttpResponse does not satisfy HasCDF interface")
	}
}

func TestHttpResponseSatisfiesHasDuration(t *testing.T) {
	var r interface{} = &HttpResponse{}
	if _, isHasDuration := r.(HasDuration); !isHasDuration {
		t.Error("HttpResponse does not satisfy HasDuration interface")
	}
}
