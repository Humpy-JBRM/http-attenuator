package data

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Recorder interface {
	SaveRequest(req *GatewayRequest) error
	SaveResponse(resp *GatewayResponse) error
}

type RecorderImpl struct {
	Requests      string `yaml:"requests" json:"requests"`
	Responses     string `yaml:"responses" json:"responses"`
	requestsChan  chan *GatewayRequest
	responsesChan chan *GatewayResponse
}

func (r *RecorderImpl) Backpatch() error {
	if r.Requests != "" {
		os.MkdirAll(r.Requests, 0755)
	}
	if r.Responses != "" {
		if r.Responses != "" {
			os.MkdirAll(r.Responses, 0755)
		}
	}

	r.requestsChan = make(chan *GatewayRequest, 100)
	r.responsesChan = make(chan *GatewayResponse, 100)
	go r.saveWorker()
	return nil
}

func (r *RecorderImpl) SaveRequest(req *GatewayRequest) error {
	if req != nil {
		r.requestsChan <- req
	}
	return nil
}

func (r *RecorderImpl) SaveResponse(resp *GatewayResponse) error {
	if resp != nil {
		r.responsesChan <- resp
	}
	return nil
}

var fileSaveWorkers = promauto.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "faultmonkey",
		Name:      "file_save_workers",
		Help:      "The number of file save workers running",
	},
)
var filesSaved = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "files_saved",
		Help:      "The number of files saved, keyed by upstream/backend and type (request/response)",
	},
	[]string{"upstream", "backend", "type"},
)
var filesSavedErrors = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "faultmonkey",
		Name:      "files_saved_errors",
		Help:      "The number of files saved, keyed by upstream/backend and type (request/response)",
	},
	[]string{"upstream", "backend", "type"},
)

func makeDirectoryHierarchy(root string, customer string, tag string, upstream string, backend string) string {
	elements := []string{root}
	if customer != "" && customer != "(unknown)" {
		elements = append(elements, customer)
	}
	if tag != "" {
		elements = append(elements, tag)
	}
	if upstream != "" {
		elements = append(elements, upstream)
	}
	if backend != "" {
		elements = append(elements, backend)
	}

	return strings.Join(elements, "/")
}

func (r *RecorderImpl) saveWorker() {
	fileSaveWorkers.Inc()
	defer func() {
		fileSaveWorkers.Dec()
	}()
	for {
		var upstream string
		var backend string
		var id string
		select {
		case req := <-r.requestsChan:
			id = req.Id
			upstream = req.Headers.Get(HEADER_X_FAULTMONKEY_UPSTREAM)
			backend = req.Headers.Get(HEADER_X_FAULTMONKEY_BACKEND)
			filesSaved.WithLabelValues(upstream, backend, "request").Inc()
			jsonBytes, err := json.MarshalIndent(req, "", "  ")
			if err != nil {
				filesSavedErrors.WithLabelValues(upstream, backend, "request").Inc()
				log.Printf("saveWorker(): request %s: %s", id, jsonBytes)
				continue
			}
			saveDir := makeDirectoryHierarchy(
				r.Requests,
				req.Headers.Get(HEADER_X_FAULTMONKEY_API_CUSTOMER),
				req.Headers.Get(HEADER_X_FAULTMONKEY_TAG),
				req.Headers.Get(HEADER_X_FAULTMONKEY_UPSTREAM),
				req.Headers.Get(HEADER_X_FAULTMONKEY_BACKEND),
			)
			os.MkdirAll(saveDir, 0755)
			filename := filepath.Join(saveDir, fmt.Sprintf("%s-request.json", id))
			err = os.WriteFile(filename, jsonBytes, 0644)
			if err != nil {
				filesSavedErrors.WithLabelValues(upstream, backend, "request").Inc()
				log.Printf("saveWorker(): request %s: %s", id, jsonBytes)
				continue
			}

		case resp := <-r.responsesChan:
			id = resp.Id
			jsonBytes, err := json.MarshalIndent(resp, "", "  ")
			if err != nil {
				log.Printf("saveWorker(): response %s: %s", id, jsonBytes)
				continue
			}
			filesSaved.WithLabelValues(upstream, backend, "response").Inc()
			saveDir := makeDirectoryHierarchy(
				r.Responses,
				resp.Headers.Get(HEADER_X_FAULTMONKEY_API_CUSTOMER),
				resp.Headers.Get(HEADER_X_FAULTMONKEY_TAG),
				resp.Headers.Get(HEADER_X_FAULTMONKEY_UPSTREAM),
				resp.Headers.Get(HEADER_X_FAULTMONKEY_BACKEND),
			)
			os.MkdirAll(saveDir, 0755)
			filename := filepath.Join(saveDir, fmt.Sprintf("%s-request.json", id))
			err = os.WriteFile(filename, jsonBytes, 0644)
			if err != nil {
				filesSavedErrors.WithLabelValues(upstream, backend, "response").Inc()
				log.Printf("saveWorker(): response %s: %s", id, jsonBytes)
				continue
			}
		}
	}
}
