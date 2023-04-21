package attenuator

import (
	"encoding/json"
	"fmt"
	"http-attenuator/util"
	"strings"
	"sync"
)

// Wrap the functions in a struct to make them much easier to test.
//
// For the purpose of this exercise, we're just over-riding the
// default HTTP function with our own, so we can very carefully
// curate how the HTTP behaves.
//
// We could equally use mocks for this, but that's too heavyweight
// for this one particular case.
type httpPostFunc func(theUrl string, payload []byte, header ...string) (int, []byte, map[string]string, error)
type httpGetFunc func(url string, header ...string) (int, []byte, map[string]string, error)
type backoffFunc func(attemptNumber int)

type httpClientWithBackoff struct {
	postFunc     httpPostFunc
	getFunc      httpGetFunc
	trafficLight string
	retries      int
}

var clientFactoryInstance *clientFactory
var clientFactoryOnce sync.Once

type clientFactory struct {
	clientImpl httpClientWithBackoff
}

func getClientFactory() *clientFactory {
	clientFactoryOnce.Do(func() {
		clientFactoryInstance = &clientFactory{
			clientImpl: *newHttpClientWithBackoff(nil, nil, "", 3),
		}
	})
	return clientFactoryInstance
}

func (f *clientFactory) Build() httpClientWithBackoff {
	return f.clientImpl
}

func (f *clientFactory) Reset() *clientFactory {
	f.clientImpl = *newHttpClientWithBackoff(nil, nil, "", 3)
	return f
}

func (f *clientFactory) SetClient(client httpClientWithBackoff) *clientFactory {
	f.clientImpl = client
	return f
}

// newHttpClientWithBackoff(nil, nil, nil, structs.MaxAttemptsSTTRequests)

func newHttpClientWithBackoff(postFunc httpPostFunc, getFunc httpGetFunc, trafficLight string, retries int) *httpClientWithBackoff {
	client := &httpClientWithBackoff{
		postFunc:     postFunc,
		getFunc:      getFunc,
		trafficLight: trafficLight,
		retries:      retries,
	}
	if client.postFunc == nil {
		client.postFunc = util.HttpPost
	}
	if client.getFunc == nil {
		client.getFunc = util.HttpGet
	}
	if client.retries <= 0 {
		client.retries = 1 // try at least once
	}

	return client
}

// ErrorPayload is used to see if we get an error message back from
// the API server.
//
// IoW ... a field called "error"
type ErrorPayload struct {
	Error interface{} `json:"error"`
}

func (c *httpClientWithBackoff) Get(url string) (code int, bodyBytes []byte, responseHeaders map[string]string, err error) {
	attemptNumber := 0
	for attemptNumber < c.retries {
		// one more attempt
		attemptNumber++

		// do the http call
		code, bodyBytes, responseHeaders, err = c.getFunc(url)
		if err != nil {
			// differentiate between 'conn reset' (try again)
			// and any other error (don't try again)
			if strings.Contains(err.Error(), "connection reset by peer") {
				// connection reset - try again
				if attemptNumber < c.retries {
					WaitForGreen(c.trafficLight, attemptNumber)
				}
				continue
			}

			/// non-transient error
			// set the error and set the terminating condition
			break
		}

		// Do we have a 'please retry' error in the response payload?
		var ep ErrorPayload
		_ = json.Unmarshal(bodyBytes, &ep)
		if strings.Contains(fmt.Sprintf("%s", ep.Error), "please retry") {
			// this is an error that whisper thinks we should retry.
			if attemptNumber < c.retries {
				WaitForGreen(c.trafficLight, attemptNumber)
			}
			continue
		}

		// we only get to here if there is no error
		break
	}

	if c.retries > 1 && attemptNumber >= c.retries {
		return 0, bodyBytes, map[string]string{}, fmt.Errorf("Get(): too many retries (%d)", c.retries)
	}

	return code, bodyBytes, responseHeaders, err
}

func (c *httpClientWithBackoff) Post(url string, payload []byte) (code int, bodyBytes []byte, responseHeaders map[string]string, err error) {
	attemptNumber := 0
	for attemptNumber < c.retries {
		// one more attempt
		attemptNumber++

		// do the http call
		code, bodyBytes, responseHeaders, err = c.postFunc(url, payload)
		if err != nil {
			// differentiate between 'conn reset' (try again)
			// and any other error (don't try again)
			if strings.Contains(err.Error(), "connection reset by peer") {
				// connection reset - try again
				if attemptNumber < c.retries {
					WaitForGreen(c.trafficLight, attemptNumber)
				}
				continue
			}

			/// non-transient error
			// set the error and set the terminating condition
			break
		}

		// Do we have a 'please retry' error in the response payload?
		var ep ErrorPayload
		_ = json.Unmarshal(bodyBytes, &ep)
		if strings.Contains(fmt.Sprintf("%s", ep.Error), "please retry") {
			// this is an error that whisper thinks we should retry.
			if attemptNumber < c.retries {
				WaitForGreen(c.trafficLight, attemptNumber)
			}
			continue
		}

		// we only get to here if there is no error
		break
	}

	if c.retries > 1 && attemptNumber >= c.retries {
		return code, bodyBytes, map[string]string{}, fmt.Errorf("Get(): too many retries (%d)", c.retries)
	}

	return code, bodyBytes, responseHeaders, err
}
