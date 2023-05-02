package data

import (
	config "http-attenuator/facade/config"
	"net/http"
	"os"
	"reflect"
	"testing"
)

func TestParseConfigYaml(t *testing.T) {
	os.Setenv("CONFIG_FILE", "../test_resources/server/pathology_config.yml")
	config.Config()

	appConfig, err := LoadConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		t.Fatal(err)
	}

	// We should have a pathology profile called 'simple'
	exists := false
	var simplePathologyProfile PathologyProfile
	if simplePathologyProfile, exists = appConfig.Config.Pathologies["simple"]; !exists {
		t.Fatal("No 'simple' pathology profile")
	}

	// This simple pathology should have a httpcode pathology and a timeout pathology
	httpcodePathology := simplePathologyProfile["httpcode"]
	if httpcodePathology == nil {
		t.Fatal("Profile'simple' does not have the expected 'httpcode' pathology")
	}

	// httpcode should have five responses
	if len(httpcodePathology.Responses) != 5 {
		t.Errorf("Expected 5 responses in httpcode pathology, but got %d", len(httpcodePathology.Responses))
	}

	// HTTP 200
	expectedWeight := 80
	expectedHeaders := http.Header{
		"Content-type": []string{"application/json"},
	}
	expectedBody := `{"success": true, "pathology": "simple", "handler": "httpcode"}`
	actualWeight := httpcodePathology.Responses[200].Weight
	actualHeaders := httpcodePathology.Responses[200].Headers
	actualBody := httpcodePathology.Responses[200].Body
	if expectedWeight != actualWeight {
		t.Errorf("httpcode.200: expected weight=%d, got %d", expectedWeight, actualWeight)
	}
	if expectedBody != actualBody {
		t.Errorf("httpcode.200: expected body='%s', got '%s'", expectedBody, actualBody)
	}
	if !reflect.DeepEqual(expectedHeaders, actualHeaders) {
		t.Errorf("httpcode.200: expected headers='%v', got '%v'", expectedHeaders, actualHeaders)
	}

	// HTTP 429
	expectedWeight = 5
	expectedHeaders = http.Header{
		"X-Backoff-Millis": []string{"60000"},
		"X-Retry-After":    []string{"now() + 60s"},
	}
	expectedBody = ""
	actualWeight = httpcodePathology.Responses[429].Weight
	actualHeaders = httpcodePathology.Responses[429].Headers
	actualBody = httpcodePathology.Responses[429].Body
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
	timeoutPathology := simplePathologyProfile["timeout"]
	if timeoutPathology == nil {
		t.Fatal("Profile'simple' does not have the expected 'timeout' pathology")
	}
	expectedWeight = 10
	actualWeight = timeoutPathology.Weight
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
	actualCode := timeoutPathology.SelectResponse().Code
	actualHeaders = timeoutPathology.SelectResponse().Headers
	actualBody = timeoutPathology.SelectResponse().Body
	if expectedCode != actualCode {
		t.Errorf("timeout: expected code=%d, got %d", expectedCode, actualCode)
	}
	if expectedBody != actualBody {
		t.Errorf("timeout: expected body='%s', got '%s'", expectedBody, actualBody)
	}
	if !reflect.DeepEqual(expectedHeaders, actualHeaders) {
		t.Errorf("timeout: expected headers='%v', got '%v'", expectedHeaders, actualHeaders)
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
	if host.Pathology != expectedPathology {
		t.Errorf("Expected server host '%s' to have pathology '%s', but it has '%s'", hostname, expectedPathology, host.Pathology)
	}
	hostname = "goodboy.com"
	expectedPathology = "good_boy"
	host = appConfig.Config.Server.Hosts[hostname]
	if host.Pathology != expectedPathology {
		t.Errorf("Expected server host '%s' to have pathology '%s', but it has '%s'", hostname, expectedPathology, host.Pathology)
	}
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
