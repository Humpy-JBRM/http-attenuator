package server

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type returnCodeCdf struct {
	code    int
	weight  int
	cdf     float64
	headers http.Header
}

type HttpCodeHandler struct {
	BaseHandler
	cdf []returnCodeCdf
	rng *rand.Rand
}

// NewHttpCodeHandlerFromConfig creates a handler from the provided config,
// which could be YML, JSON or just a map
func NewHttpCodeHandlerFromConfig(map[string]interface{}) (Handler, error) {
	return nil, nil
}

func NewHttpCodeHandler(name string, weights map[int]int) Handler {
	cdf := make([]returnCodeCdf, 0)

	// Calculate the total weight
	totalWeight := 0
	for code, weight := range weights {
		totalWeight += weight
		rcdf := returnCodeCdf{
			code:    code,
			weight:  weight,
			headers: make(http.Header),
		}
		cdf = append(cdf, rcdf)
	}

	// Backpatch the cdf
	for i := 0; i < len(cdf); i++ {
		cdf[i].cdf = float64(float64(cdf[i].weight) / float64(totalWeight))
	}

	return &HttpCodeHandler{
		BaseHandler: BaseHandler{
			name: name,
		},
		cdf: cdf,
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// HttpCodeHandler returns a variety of status codes according
// to the cdf
func (h *HttpCodeHandler) Handle(c *gin.Context) {
	probability := h.rng.Float64()
	for _, cumulative := range h.cdf {
		if cumulative.cdf <= probability {
			c.Status(cumulative.code)
			return
		}
	}

	// Default to HTTP 200 OK
	c.Status(http.StatusOK)
}
