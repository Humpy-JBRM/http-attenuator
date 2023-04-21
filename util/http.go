package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func HttpGet(url string, header ...string) (int, []byte, map[string]string, error) {
	var netClient = &http.Client{
		Timeout: time.Second * 60,
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return http.StatusBadRequest, []byte{}, map[string]string{}, err
	}

	// Add the headers
	for _, h := range header {
		fields := strings.Split(strings.TrimSpace(h), ":")
		if len(fields) != 2 {
			return http.StatusBadRequest, []byte{}, map[string]string{}, fmt.Errorf("Invalid header: %s", h)
		}

		request.Header.Add(strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1]))
	}

	response, err := netClient.Do(request)
	if err != nil {
		return http.StatusBadRequest, []byte{}, map[string]string{}, err
	}
	if response == nil {
		return http.StatusBadRequest, []byte{}, map[string]string{}, fmt.Errorf("ERROR: %s: Got nil response from server", url)
	}

	responseHeaders := make(map[string]string)
	for k, v := range response.Header {
		responseHeaders[k] = v[0]
	}
	responseBytes, err := ioutil.ReadAll(response.Body)
	response.Body.Close()

	return response.StatusCode, responseBytes, responseHeaders, err
}

// Do NOT follow redirects
func doNotRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

func HttpPost(theUrl string, payload []byte, header ...string) (int, []byte, map[string]string, error) {
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
		return 0, []byte{}, map[string]string{}, err
	}

	// Add the headers
	for _, h := range header {
		fields := strings.Split(strings.TrimSpace(h), ":")
		if len(fields) != 2 {
			return 0, []byte{}, map[string]string{}, fmt.Errorf("Invalid header: %s", h)
		}

		request.Header.Add(strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1]))
	}

	response, err := netClient.Do(request)
	if err != nil {
		code := 0
		if response != nil {
			code = response.StatusCode
		}
		return code, []byte{}, map[string]string{}, err
	}
	if response == nil {
		return 0, []byte{}, map[string]string{}, fmt.Errorf("ERROR: %s: Got nil response from server", theUrl)
	}

	responseBytes, err := ioutil.ReadAll(response.Body)
	response.Body.Close()

	responseHeaders := make(map[string]string)
	for k, v := range response.Header {
		responseHeaders[k] = v[0]
	}
	response.Body.Close()

	return response.StatusCode, responseBytes, responseHeaders, err
}

func HttpPut(theUrl string, payload []byte, header ...string) (int, []byte, error) {
	var netClient = &http.Client{
		// TODO(john): cookies
		// Jar:     GCookieJar,
		Timeout:       time.Second * 300,
		CheckRedirect: doNotRedirect,
	}

	// Authenticate
	var request *http.Request
	var err error
	request, err = http.NewRequest("PUT", theUrl, bytes.NewBuffer(payload))
	if err != nil {
		return 0, []byte{}, err
	}

	// Add the headers
	for _, h := range header {
		fields := strings.Split(strings.TrimSpace(h), ":")
		if len(fields) != 2 {
			return 0, []byte{}, fmt.Errorf("Invalid header: %s", h)
		}

		request.Header.Add(strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1]))
	}

	response, err := netClient.Do(request)
	if err != nil {
		code := 0
		if response != nil {
			code = response.StatusCode
		}
		return code, []byte{}, err
	}
	if response == nil {
		return 0, []byte{}, fmt.Errorf("ERROR: %s: Got nil response from server", theUrl)
	}

	responseBytes, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	// log.Println(string(GResponseBytes))
	return response.StatusCode, responseBytes, err
}

func HttpDelete(url string, header ...string) (int, []byte, error) {
	var netClient = &http.Client{
		Timeout: time.Second * 60,
	}

	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return http.StatusBadRequest, []byte{}, err
	}

	// Add the headers
	for _, h := range header {
		fields := strings.Split(strings.TrimSpace(h), ":")
		if len(fields) != 2 {
			return http.StatusBadRequest, []byte{}, fmt.Errorf("Invalid header: %s", h)
		}

		request.Header.Add(strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1]))
	}

	response, err := netClient.Do(request)
	if err != nil {
		return http.StatusBadRequest, []byte{}, err
	}
	if response == nil {
		return http.StatusBadRequest, []byte{}, fmt.Errorf("ERROR: %s: Got nil response from server", url)
	}

	responseBytes, err := ioutil.ReadAll(response.Body)
	response.Body.Close()

	return response.StatusCode, responseBytes, err
}
