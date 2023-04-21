package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func HttpGet(url string, headers http.Header) (int, []byte, http.Header, error) {
	var netClient = &http.Client{
		Timeout: time.Second * 60,
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return http.StatusBadRequest, []byte{}, http.Header{}, err
	}
	request.Header = headers

	response, err := netClient.Do(request)
	if err != nil {
		return http.StatusBadRequest, []byte{}, http.Header{}, err
	}
	if response == nil {
		return http.StatusBadRequest, []byte{}, http.Header{}, fmt.Errorf("ERROR: %s: Got nil response from server", url)
	}

	responseBytes, err := ioutil.ReadAll(response.Body)
	response.Body.Close()

	return response.StatusCode, responseBytes, response.Header, err
}

// Do NOT follow redirects
func doNotRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

func HttpPost(theUrl string, payload []byte, headers http.Header) (int, []byte, http.Header, error) {
	var netClient = &http.Client{
		// TODO(john): cookies
		// Jar:     GCookieJar,
		Timeout:       time.Second * 60,
		CheckRedirect: doNotRedirect,
	}

	// Authenticate
	var request *http.Request
	var err error
	request, err = http.NewRequest("POST", theUrl, bytes.NewBuffer(payload))
	if err != nil {
		return 0, []byte{}, http.Header{}, err
	}
	request.Header = headers

	response, err := netClient.Do(request)
	if err != nil {
		code := 0
		if response != nil {
			code = response.StatusCode
		}
		return code, []byte{}, http.Header{}, err
	}
	if response == nil {
		return 0, []byte{}, http.Header{}, fmt.Errorf("ERROR: %s: Got nil response from server", theUrl)
	}

	responseBytes, err := ioutil.ReadAll(response.Body)
	response.Body.Close()

	response.Body.Close()

	return response.StatusCode, responseBytes, headers, err
}
