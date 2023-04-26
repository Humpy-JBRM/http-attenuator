package broker

import (
	"http-attenuator/data"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestBackendForwardProxy(t *testing.T) {
	os.Setenv("CONFIG_FILE", "../test_resources/broker/config.yml")
	serviceMap := GetServiceMap()
	if serviceMap == nil {
		t.Fatal("ServiceMap from config was nil")
	}

	// Get the bing backend
	preferred := "bing"
	backend := GetServiceMap().GetBackend("search", "bing")
	if backend == nil {
		t.Fatalf("Could not get preferred backend '%s'", preferred)
	}
	if !strings.EqualFold(preferred, backend.Label) {
		t.Fatalf("Wanted preferred backend '%s', but got '%s'", preferred, backend.Label)
	}

	// Create the request
	//
	// TODO(john): all of this needs encapsulated away in the DoSync()
	//
	// Map any query params
	myQuery := url.Values{
		"q": []string{"wibble"},
	}
	searchUrl := backend.Url
	queryToSend := searchUrl.Query()
	for k, v := range myQuery {
		// Are we mapping this query parameter from one name to another?
		if mapped, isMapped := backend.Params[k]; isMapped {
			queryToSend.Set(mapped, v[0])
			continue
		}

		// No query name mapping
		queryToSend.Set(k, v[0])
	}
	searchUrl.RawQuery = queryToSend.Encode()

	req := data.NewGatewayRequest(
		"",
		"GET",
		searchUrl,
		http.Header{},
		nil,
	)

	// Now map any headers
	//
	// TODO(john): encapsulate this
	for k, values := range backend.Headers {
		for _, v := range values {
			req.Headers.Add(k, v)
		}
	}

	resp, err := NewForwardProxy(backend).DoSync(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("Got nil response")
	}
	t.Fatal("")
}
