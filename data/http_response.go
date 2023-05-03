package data

import (
	"net/http"
	"time"
)

type HttpResponse struct {
	// This needs to be backpatched in the case of a httpcode,
	// because it lives in a map
	Code     int         `yaml:"code" json:"code"`
	Weight   int         `yaml:"weight" json:"weight"`
	Duration string      `yaml:"duration" json:"duration"`
	Headers  http.Header `yaml:"headers" json:"headers"`
	Body     string      `yaml:"body" json:"body"`

	// this needs to be backpatched because it is derived
	// from the config value (which could be a formula)
	durationConfig HasDuration

	// this needs to be backpatched so we can select responses
	// according to a cdf
	cdf float64
}

// HttpResponse should satisfy the HasDuration interface
func (r *HttpResponse) GetDuration() *time.Duration {
	if r.durationConfig == nil {
		return nil
	}

	// delegate to the duration config to give us a duration
	// according to its distribution
	return r.durationConfig.GetDuration()
}

// HttpResponse should satisfy the HasCDF interface
func (r *HttpResponse) CDF() float64 {
	return r.cdf
}

// HttpResponse should satisfy the HasCDF interface
func (r *HttpResponse) SetCDF(cdf float64) {
	r.cdf = cdf
}

// HttpResponse should satisfy the HasCDF interface
func (r *HttpResponse) GetWeight() int {
	return r.Weight
}
