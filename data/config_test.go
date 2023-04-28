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
	if simplePathologyProfile, exists = appConfig.Config.Profiles["simple"]; !exists {
		t.Fatal("No 'simple' pathology profile")
	}

	// This simple pathology should have a httpcode pathology and a timeout pathology
	httpcodePathology := simplePathologyProfile.HttpCode
	if httpcodePathology == nil {
		t.Fatal("Profile'simple' does not have the expected 'httpcode' pathology")
	}

	// httpcode should have five responses
	if len(httpcodePathology.Responses) != 5 {
		t.Errorf("Expected 5 responses in httpcode pathology, but got %d", len(httpcodePathology.Responses))
	}

	expectedWeight := 80
	expectedHeaders := http.Header{
		"Content-type": []string{"application/json"},
	}
	expectedBody := `{"success": true}`
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

	// timeout pathology
	timeoutPathology := simplePathologyProfile.Timeout
	if timeoutPathology == nil {
		t.Fatal("Profile'simple' does not have the expected 'timeout' pathology")
	}
	expectedWeight = 10
	actualWeight = timeoutPathology.Weight
	if expectedWeight != actualWeight {
		t.Errorf("timeout: expected weight=%d, got %d", expectedWeight, actualWeight)
	}

	// The timeout duration
	expectedMillis := int64(10000)
	actualMillis := timeoutPathology.Millis
	if expectedMillis != actualMillis {
		t.Errorf("timeout: expected weight=%d, got %d", expectedMillis, actualMillis)
	}

	// The timeout response
	expectedCode := 200
	expectedHeaders = http.Header{
		"Content-type": []string{"application/json"},
	}
	expectedBody = `{"success": true}`
	actualCode := timeoutPathology.Response.Code
	actualHeaders = timeoutPathology.Response.Headers
	actualBody = timeoutPathology.Response.Body
	if expectedCode != actualCode {
		t.Errorf("timeout: expected code=%d, got %d", expectedCode, actualCode)
	}
	if expectedBody != actualBody {
		t.Errorf("timeout: expected body='%s', got '%s'", expectedBody, actualBody)
	}
	if !reflect.DeepEqual(expectedHeaders, actualHeaders) {
		t.Errorf("timeout: expected headers='%v', got '%v'", expectedHeaders, actualHeaders)
	}
}
