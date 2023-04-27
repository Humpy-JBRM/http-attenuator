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

type HttpCodeCdf struct {
	code           int
	weight         int
	cdf            float64
	durationMean   float64
	durationStddev float64
	headers        http.Header
	jsonBody       []byte
}

type HttpCodeCdfBuilder interface {
	Reset() HttpCodeCdfBuilder
	Code(code int) HttpCodeCdfBuilder
	Weight(weight int) HttpCodeCdfBuilder
	CDF(cdf float64) HttpCodeCdfBuilder
	DurationMean(mean float64) HttpCodeCdfBuilder
	DurationStddev(stddev float64) HttpCodeCdfBuilder
	AddHeader(name string, value string) HttpCodeCdfBuilder
	Body(jsonBytes []byte) HttpCodeCdfBuilder
	Build() *HttpCodeCdf
}

type HttpCodeCdfBuilderImpl struct {
	impl HttpCodeCdf
}

func NewHttpCodeCdfBuilder() HttpCodeCdfBuilder {
	return &HttpCodeCdfBuilderImpl{
		impl: HttpCodeCdf{
			headers: make(http.Header),
		},
	}
}

func (b *HttpCodeCdfBuilderImpl) Reset() HttpCodeCdfBuilder {
	b.impl = HttpCodeCdf{
		headers: make(http.Header),
	}
	return b
}

func (b *HttpCodeCdfBuilderImpl) Code(code int) HttpCodeCdfBuilder {
	b.impl.code = code
	return b
}

func (b *HttpCodeCdfBuilderImpl) Weight(weight int) HttpCodeCdfBuilder {
	b.impl.weight = weight
	return b
}

func (b *HttpCodeCdfBuilderImpl) CDF(cdf float64) HttpCodeCdfBuilder {
	b.impl.cdf = cdf
	return b
}

func (b *HttpCodeCdfBuilderImpl) DurationMean(mean float64) HttpCodeCdfBuilder {
	b.impl.durationMean = mean
	return b
}

func (b *HttpCodeCdfBuilderImpl) DurationStddev(stddev float64) HttpCodeCdfBuilder {
	b.impl.durationStddev = stddev
	return b
}

func (b *HttpCodeCdfBuilderImpl) AddHeader(name string, value string) HttpCodeCdfBuilder {
	b.impl.headers.Add(name, value)
	return b
}

func (b *HttpCodeCdfBuilderImpl) Body(jsonBytes []byte) HttpCodeCdfBuilder {
	b.impl.jsonBody = jsonBytes
	return b
}

func (b *HttpCodeCdfBuilderImpl) Build() *HttpCodeCdf {
	defensiveCopy := b.impl
	return &defensiveCopy
}

// make returnCodeCdf conform to the HasCDF interface
// so we can use our generic ChooseFromCDF() function
func (f *HttpCodeCdf) CDF() float64 {
	return f.cdf
}

type HttpCodeHandler struct {
	BaseHandler
	cdf []data.HasCDF
	rng *rand.Rand
}

// HttpCodeHandler returns a variety of status codes according
// to the cdf
func (h *HttpCodeHandler) Handle(c *gin.Context) {
	// get the pathology - this is just for logging / monitoring
	pathology, _ := c.Get("pathology")
	pathologyRequests.WithLabelValues(fmt.Sprint(pathology), h.name, c.Request.Method).Inc()
	choice := data.ChooseFromCDF(h.rng.Float64(), h.cdf)
	if choice == nil {
		pathologyErrors.WithLabelValues(fmt.Sprint(pathology), h.name, c.Request.Method).Inc()
		err := fmt.Errorf("%s.Handle(): could not get handler", h.name)
		log.Println(err)
		c.AbortWithError(
			http.StatusInternalServerError,
			err,
		)
		return
	}

	// TODO(john): headers
	pathologyResponses.WithLabelValues(fmt.Sprint(pathology), h.name, c.Request.Method, fmt.Sprint(choice.(*HttpCodeCdf).code)).Inc()

	// If there is a delay / duration distribution, then do the sleep.
	//
	// The value is in seconds
	if choice.(*HttpCodeCdf).durationMean > 0 && choice.(*HttpCodeCdf).durationStddev > 0 {
		sleepSeconds := rand.NormFloat64()*choice.(*HttpCodeCdf).durationStddev + choice.(*HttpCodeCdf).durationMean
		if sleepSeconds > 0 {
			sleepTime := time.Duration(int64(sleepSeconds*1000)) * time.Millisecond
			pathologyLatency.WithLabelValues(fmt.Sprint(pathology), h.name, c.Request.Method).Add(sleepSeconds * 1000)
			time.Sleep(sleepTime)
		}
	}

	c.Status(choice.(*HttpCodeCdf).code)
}
