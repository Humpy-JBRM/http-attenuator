package data

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler interface {
	GetName() string
	Handle(c *gin.Context)
}

type Pathology interface {
	Handler
	HasCDF
	GetProfileName() string
	SelectResponse() *HttpResponse
}

type PathologyImpl struct {
	Weight    int                   `yaml:"weight" json:"weight"`
	Duration  string                `yaml:"duration" json:"duration"`
	Responses map[int]*HttpResponse `yaml:"responses" json:"responses"`

	// The CDF when this pathology is part of a profile
	cdf float64

	// These get backpatched in LoadConfig()
	name              string
	profile           string
	rng               *rand.Rand
	responsesAsHasCDF []HasCDF
}

func (p *PathologyImpl) GetName() string {
	return p.name
}

func (p *PathologyImpl) GetProfileName() string {
	return p.profile
}

// PathologyImpl must conform to Pathology (HasCDF) duck-type
func (p *PathologyImpl) CDF() float64 {
	return p.cdf
}

func (p *PathologyImpl) SetCDF(cdf float64) {
	p.cdf = cdf
}

func (p *PathologyImpl) GetWeight() int {
	return p.Weight
}

// SelectResponse selects the HttpResponse to be returned
// based on the cdf
func (p *PathologyImpl) SelectResponse() *HttpResponse {
	if len(p.Responses) == 0 {
		return nil
	}
	for _, resp := range p.Responses {
		return resp
	}

	if p.responsesAsHasCDF == nil {
		return nil
	}
	return ChooseFromCDF(p.rng.Float64(), p.responsesAsHasCDF).(*HttpResponse)
}

// Satisfy the Handler duck type
func (p *PathologyImpl) Handle(c *gin.Context) {
	resp := p.SelectResponse()
	pathologyRequests.WithLabelValues(
		p.profile,                       // profile
		p.name,                          // pathology
		strings.ToLower(c.Request.Host), //host
		c.Request.Method,                // method
		fmt.Sprint(resp.Code),           // code
	).Inc()

	now := time.Now().UTC().UnixMilli()
	defer func(now int64) {
		pathologyLatency.WithLabelValues(
			p.profile,                       // profile
			p.name,                          // pathology
			strings.ToLower(c.Request.Host), //host
			c.Request.Method,                // method
			fmt.Sprint(resp.Code),           // code
		).Add(float64(time.Now().UTC().UnixMilli() - now))
		pathologyResponses.WithLabelValues(
			p.profile,                       // profile
			p.name,                          // pathology
			strings.ToLower(c.Request.Host), //host
			c.Request.Method,                // method
			fmt.Sprint(resp.Code),           // code
		).Inc()
	}(now)

	if resp == nil {
		log.Printf("%s.Handle(%s): no response configured", p.name, c.Request.URL.String())
		pathologyErrors.WithLabelValues(
			p.profile,                       // profile
			p.name,                          // pathology
			strings.ToLower(c.Request.Host), //host
			c.Request.Method,                // method
			"",                              // code
		).Inc()
		return
	}

	// delay for the configured amount of time
	if resp.GetDuration() != nil && resp.GetDuration().Milliseconds() > 0 {
		time.Sleep(*resp.GetDuration())
	}

	// Response code
	c.Status(resp.Code)

	// Headers
	for headerName, values := range resp.Headers {
		for _, value := range values {
			c.Writer.Header().Add(headerName, value)
		}
	}

	// Response body
	c.Writer.Write([]byte(resp.Body))
}
