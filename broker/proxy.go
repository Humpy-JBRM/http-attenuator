package broker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"http-attenuator/data"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// ForwardProxy is a simple forward proxy which just plumbs a request
// directly to a backend
type ForwardProxy struct {
	backend *Backend
}

func NewForwardProxy(backend *Backend) *ForwardProxy {
	return &ForwardProxy{
		backend: backend,
	}
}

func (p *ForwardProxy) DoSync(req *data.GatewayRequest) (response *data.GatewayResponse, err error) {
	now := time.Now().UTC().UnixMilli()
	var recordRequestFile *os.File
	if p.backend.RecordRequestRoot != "" {
		recordPath := filepath.Join(p.backend.RecordRequestRoot, req.GetUrl().Host, req.Id+"-request.json")
		os.MkdirAll(filepath.Dir(recordPath), 0755)
		recordRequestFile, err = os.OpenFile(recordPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			err = fmt.Errorf("ForwardProxy.DoSync(%s): %s", req.GetUrl().String(), err.Error())
			log.Println(err)
			response.Headers.Add(data.HEADER_X_ATTENUATOR_ERROR, err.Error())
			return response, err
		}
		defer recordRequestFile.Close()
	}
	var recordResponseFile *os.File
	if p.backend.RecordResponseRoot != "" {
		recordPath := filepath.Join(p.backend.RecordResponseRoot, req.GetUrl().Host, req.Id+"-response.json")
		os.MkdirAll(filepath.Dir(recordPath), 0755)
		recordResponseFile, err = os.OpenFile(recordPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			err = fmt.Errorf("ForwardProxy.DoSync(%s): %s", req.GetUrl().String(), err.Error())
			log.Println(err)
			response.Headers.Add(data.HEADER_X_ATTENUATOR_ERROR, err.Error())
			return response, err
		}
		defer recordResponseFile.Close()
	}
	if recordRequestFile != nil {
		if req.WhenMillis == 0 {
			req.WhenMillis = time.Now().UTC().UnixMilli()
		}
		reqJson, err := json.MarshalIndent(req, "", "  ")
		if err != nil {
			panic(err)
		}
		_, err = recordRequestFile.Write(reqJson)
		if err != nil {
			panic(err)
		}
	}
	response = &data.GatewayResponse{
		Id:             req.Id,
		Headers:        req.Headers,
		Body:           req.Body,
		WhenMillis:     now,
		DurationMillis: time.Now().UTC().UnixMilli() - now,
		StatusCode:     http.StatusNotImplemented,
	}

	httpRequest := &http.Request{
		Method: req.Method,
		URL:    req.GetUrl(),
		Host:   req.GetUrl().Host,
		Header: req.Headers,
	}

	if req.Body != nil && len(req.Body) > 0 {
		httpRequest.Body = io.NopCloser(bytes.NewReader(req.Body))
		httpRequest.ContentLength = int64(len(req.Body))
	}

	// Make the request
	//
	// TODO(john): put it through the attenuator / circuit breaker etc
	client := http.Client{
		// TODO(john): any client connection settings (e.g. timeout)
	}
	log.Printf("ForwardProxy(): %s", httpRequest.URL.String())
	resp, e := client.Do(httpRequest)
	if err != nil {
		err = fmt.Errorf("ForwardProxy.DoSync(%s): %s", req.GetUrl().String(), e.Error())
		log.Println(err)
		response.Headers.Add(data.HEADER_X_ATTENUATOR_ERROR, err.Error())
		return response, err
	}

	defer resp.Body.Close()

	// Populate the response
	response.StatusCode = resp.StatusCode
	response.Headers = resp.Header
	response.Body = []byte{}
	response.Body, err = io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("ForwardProxy.DoSync(%s): %s", req.GetUrl().String(), err.Error())
		log.Println(err)
		response.Headers.Add(data.HEADER_X_ATTENUATOR_ERROR, err.Error())
	}

	if recordResponseFile != nil {
		response.WhenMillis = time.Now().UTC().UnixMilli()
		response.DurationMillis = response.WhenMillis - req.WhenMillis
		respJson, _ := json.MarshalIndent(response, "", "  ")
		recordResponseFile.Write(respJson)
	}
	return response, err
}
