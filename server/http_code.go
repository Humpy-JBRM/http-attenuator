package server

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type returnCodeCdf struct {
	code   int
	weight int
	cdf    float64
}

type HttpCodeHandler struct {
	BaseHandler
	cdf []returnCodeCdf
}

func NewHttpCodeHandler(name string, weights map[int]int) Handler {
	cdf := make([]returnCodeCdf, 0)

	// Calculate the total weight
	totalWeight := 0
	for code, weight := range weights {
		totalWeight += weight
		rcdf := returnCodeCdf{
			code:   code,
			weight: weight,
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
	}
}

// HttpCodeHandler sleeps for a given time before responding,
// or it sleeps forever and never returns.
//
// It is used to simulate servers that are slow or which
// otherwise time out
func (h *HttpCodeHandler) Handle(c *gin.Context) {
	if h.httpcodeMillis <= 0 {
		// Sleep forever.  Never returns
		log.Printf("%s.Handle(): sleep forever", h.name)
		select {}
	}

	log.Printf("%s.Handle(): sleep for %dms", h.name, h.httpcodeMillis)
	time.Sleep(time.Duration(h.httpcodeMillis))

	// TODO(john): set the response to return in config
	c.Status(http.StatusOK)
}
