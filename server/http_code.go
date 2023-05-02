package server

import (
	"fmt"
	"http-attenuator/data"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HttpCodePathology struct {
	BasePathology
	responses []*HttpCodeResponse
	cdf       []data.HasCDF
	rng       *rand.Rand
}

type HttpCodePathologyBuilder struct {
	BasePathologyBuilder
	hp HttpCodePathology
}

func NewHttpCodePathologyBuilder() *HttpCodePathologyBuilder {
	bpb := NewBasePathologyBuilder()
	return &HttpCodePathologyBuilder{
		BasePathologyBuilder: *bpb,
		hp: HttpCodePathology{
			rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
			responses: make([]*HttpCodeResponse, 0),
		},
	}
}

func (b *HttpCodePathologyBuilder) Build() (*HttpCodePathology, error) {
	// Populate the CDF
	result := b.hp
	baseBuilder := &b.BasePathologyBuilder
	base, _ := baseBuilder.Build()
	result.BasePathology = *base
	return &result, nil
}

// HttpCodeHandler returns a variety of status codes according
// to the cdf
func (h *HttpCodePathology) Handle(c *gin.Context) {
	// get the pathology - this is just for logging / monitoring
	pathologyRequests.WithLabelValues(h.profile, h.GetName(), c.Request.Method).Inc()
	choice := data.ChooseFromCDF(h.rng.Float64(), h.cdf)
	if choice == nil {
		pathologyErrors.WithLabelValues(h.profile, h.GetName(), c.Request.Method).Inc()
		err := fmt.Errorf("%s.Handle(): could not get handler", h.GetName())
		log.Println(err)
		c.AbortWithError(
			http.StatusInternalServerError,
			err,
		)
		return
	}

	httpCode := choice.(*HttpCodeResponse)

	// TODO(john): headers
	pathologyResponses.WithLabelValues(h.profile, h.GetName(), c.Request.Method, fmt.Sprint(choice.(*HttpCodeResponse).Code)).Inc()

	sleepDuration := h.GetDuration()

	// If there is a delay / duration distribution in the response
	// that was chosen, then do the sleep.
	//
	// Otherwise,
	// The value is in seconds
	if sleepDuration == nil {
		// No default duration.
		// See if the response we chose has a duration set
		if _, isHasDuration := choice.(data.HasDuration); isHasDuration {
			sleepDuration = choice.(data.HasDuration).GetDuration()
		}
	}
	if sleepDuration != nil {
		pathologyLatency.WithLabelValues(h.profile, h.GetName(), c.Request.Method).Add(float64(sleepDuration.Milliseconds()))
		time.Sleep(*sleepDuration)
	}

	// Status code
	c.Status(httpCode.Code)

	// Any headers
	//for name, values := range c.

	// Any response body
}
