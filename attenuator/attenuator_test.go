package attenuator

import (
	"http-attenuator/data"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSimpleGetSync(t *testing.T) {
	os.Setenv("CONFIG_FILE", "../../config.yml")
	hertz := 1.0
	iterations := 10
	at, err := NewAttenuator(
		"foo",
		hertz,
		0,
		1,
	)
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now().UTC().UnixMilli()
	url, _ := url.Parse("https://www.google.com")

	count := 0
	for i := 0; i < iterations; i++ {
		count++
		// Heavy lifting
		attenuatedRequest := data.NewGatewayRequest(
			"",
			"GET",
			url,
			http.Header{},
			nil,
		)
		response, err := at.DoSync(attenuatedRequest)
		if err != nil {
			t.Error(err)
		}
		// End of heavy lifting

		// check response is 200
		if response.StatusCode != http.StatusOK {
			t.Errorf("Expected response %d, got %d", http.StatusOK, response.StatusCode)
		}
		needle := ">Google.co.uk<"
		if response.Body != nil {
			haystack := string(*response.Body)
			if !strings.Contains(haystack, needle) {
				t.Errorf("Expected body to contain '%s', but got body\n%s", needle, haystack)
			}
		} else {
			t.Errorf("Expected body to contain '%s', but body is nil", needle)
		}
	}
	end := time.Now().UTC().UnixMilli()
	expectedDuration := int64(1000 * float64(iterations) / hertz)
	slop := expectedDuration / 5
	actualDuration := (end - start)
	if (actualDuration+slop) < expectedDuration || (actualDuration-slop) > expectedDuration {
		t.Fatalf("Expected %d times in ~%dms, but was %d times in %d seconds", iterations, expectedDuration, count, (end-start)/1000)
	}
}

// func TestSimplePostSync(t *testing.T) {
// os.Setenv("CONFIG_FILE", "../config.yml")
// 	iterations := 10
// 	hertz := 2.0
// 	at, err := NewAttenuator(
// 		"foo",
// 		hertz,
// 		0,
// 		1,
// 	)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	start := time.Now().UTC().UnixMilli()
// 	url, _ := url.Parse("https://www.google.com")

// 	// THE HEAVY LIFTING
// 	req := http.Request{
// 		Method: "POST",
// 		URL:    url,
// 	}
// 	attenuatedRequest := data.NewHttpRequest("", &req)
// 	// response, err := at.DoSync(attenuatedRequest)
// 	// END OF HEAVY LIFTING
// 	count := 0
// 	for i := 0; i < iterations; i++ {
// 		count++
// 		response, err := at.DoSync(attenuatedRequest)
// 		if err != nil {
// 			t.Error(err)
// 		}

// 		// check response is 200
// 		if response.Code != http.StatusMethodNotAllowed {
// 			t.Errorf("Expected response %d, got %d", http.StatusOK, response.Code)
// 		}
// 		haystack := string(response.Body)
// 		needle := "<code>POST</code> is inappropriate for the URL"
// 		if !strings.Contains(haystack, needle) {
// 			t.Errorf("Expected body to contain '%s', but got body\n%s", needle, haystack)
// 		}
// 	}
// 	end := time.Now().UTC().UnixMilli()
// 	expectedDuration := int64(1000 * float64(iterations) / hertz)
// 	slop := expectedDuration / 5
// 	actualDuration := (end - start)
// 	if (actualDuration+slop) < expectedDuration || (actualDuration-slop) > expectedDuration {
// 		t.Fatalf("Expected %d times in ~%dms, but was %d times in %d seconds", iterations, expectedDuration, count, (end-start)/1000)
// 	}
// }

// func TestSimplePostSync405IsBad(t *testing.T) {
// os.Setenv("CONFIG_FILE", "../config.yml")
// 	iterations := 10
// 	hertz := 2.0
// 	at, err := NewAttenuator(
// 		"foo",
// 		hertz,
// 		0,
// 		1,
// 	)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	start := time.Now().UTC().UnixMilli()
// 	url, _ := url.Parse("https://www.google.com")

// 	// THE HEAVY LIFTING
// 	req := http.Request{
// 		Method: "POST",
// 		URL:    url,
// 	}
// 	attenuatedRequest := data.NewHttpRequest(
// 		"",
// 		&req,
// 		func(resp *http.Response) (bool, bool) {
// 			if resp.StatusCode == 405 {
// 				// consider this successful
// 				return true, false
// 			}

// 			return false, true
// 		},
// 	)
// 	// response, err := at.DoSync(attenuatedRequest)
// 	// END OF HEAVY LIFTING
// 	count := 0
// 	for i := 0; i < iterations; i++ {
// 		count++
// 		response, err := at.DoSync(attenuatedRequest)
// 		if err != nil {
// 			t.Error(err)
// 		}

// 		// check response is 200
// 		if response.Code != http.StatusMethodNotAllowed {
// 			t.Errorf("Expected response %d, got %d", http.StatusOK, response.Code)
// 		}
// 		haystack := string(response.Body)
// 		needle := "<code>POST</code> is inappropriate for the URL"
// 		if !strings.Contains(haystack, needle) {
// 			t.Errorf("Expected body to contain '%s', but got body\n%s", needle, haystack)
// 		}
// 	}
// 	end := time.Now().UTC().UnixMilli()
// 	expectedDuration := int64(1000 * float64(iterations) / hertz)
// 	slop := expectedDuration / 5
// 	actualDuration := (end - start)
// 	if (actualDuration+slop) < expectedDuration || (actualDuration-slop) > expectedDuration {
// 		t.Fatalf("Expected %d times in ~%dms, but was %d times in %d seconds", iterations, expectedDuration, count, (end-start)/1000)
// 	}
// }
