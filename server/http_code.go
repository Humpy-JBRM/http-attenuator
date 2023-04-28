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

type HttpCodeHandler struct {
	data.BaseHandler
	cdf []data.HasCDF
	rng *rand.Rand
}

// HttpCodeHandler returns a variety of status codes according
// to the cdf
func (h *HttpCodeHandler) Handle(c *gin.Context) {
	// get the pathology - this is just for logging / monitoring
	pathologyRequests.WithLabelValues(fmt.Sprint(h.GetProfile()), h.GetName(), c.Request.Method).Inc()
	choice := data.ChooseFromCDF(h.rng.Float64(), h.cdf)
	if choice == nil {
		pathologyErrors.WithLabelValues(fmt.Sprint(h.GetProfile()), h.GetName(), c.Request.Method).Inc()
		err := fmt.Errorf("%s.Handle(): could not get handler", h.GetName())
		log.Println(err)
		c.AbortWithError(
			http.StatusInternalServerError,
			err,
		)
		return
	}

	httpCode := choice.(*data.HttpCodeCdf)

	// TODO(john): headers
	pathologyResponses.WithLabelValues(fmt.Sprint(h.GetProfile()), h.GetName(), c.Request.Method, fmt.Sprint(choice.(*data.HttpCodeCdf).Code)).Inc()

	// If there is a delay / duration distribution, then do the sleep.
	//
	// The value is in seconds
	if _, isHasDuration := choice.(data.HasDuration); isHasDuration {
		sleepDuration := choice.(data.HasDuration).GetDuration()
		if sleepDuration != nil {
			pathologyLatency.WithLabelValues(fmt.Sprint(h.GetProfile()), h.GetName(), c.Request.Method).Add(float64(sleepDuration.Milliseconds()))
			time.Sleep(*sleepDuration)
		}
	}

	// Status code
	c.Status(httpCode.Code)

	// Any headers
	//for name, values := range c.

	// Any response body
}
