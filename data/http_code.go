package data

import (
	"math/rand"
	"net/http"
	"time"
)

type HttpCodeCdf struct {
	Code           int
	Weight         int
	Cdf            float64
	DurationMean   float64
	DurationStddev float64
	DurationMillis int64
	Headers        http.Header
	JsonBody       []byte
}

// make returnCodeCdf conform to the HasCDF interface
// so we can use our generic ChooseFromCDF() function
func (c *HttpCodeCdf) CDF() float64 {
	return c.Cdf
}

func (c *HttpCodeCdf) SetCDF(cdf float64) {
	c.Cdf = cdf
}

func (c *HttpCodeCdf) GetWeight() int {
	return c.Weight
}

// GetDuration calculates a sleep time based on the
// distribution
//
// This is to conform to the HasDuration interface
func (c *HttpCodeCdf) GetDuration() *time.Duration {
	if c.DurationMean > 0 && c.DurationStddev > 0 {
		durationSeconds := rand.NormFloat64()*c.DurationStddev + c.DurationMean
		if durationSeconds > 0 {
			duration := time.Duration(int64(durationSeconds*1000)) * time.Millisecond
			return &duration
			//pathologyLatency.WithLabelValues(fmt.Sprint(pathology), h.name, c.Request.Method).Add(sleepSeconds * 1000)
		}

		return nil
	}

	if c.DurationMillis > 0 {
		duration := time.Duration(c.DurationMillis) * time.Millisecond
		return &duration
	}

	return nil
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
			Headers: make(http.Header),
		},
	}
}

func (b *HttpCodeCdfBuilderImpl) Reset() HttpCodeCdfBuilder {
	b.impl = HttpCodeCdf{
		Headers: make(http.Header),
	}
	return b
}

func (b *HttpCodeCdfBuilderImpl) Code(code int) HttpCodeCdfBuilder {
	b.impl.Code = code
	return b
}

func (b *HttpCodeCdfBuilderImpl) Weight(weight int) HttpCodeCdfBuilder {
	b.impl.Weight = weight
	return b
}

func (b *HttpCodeCdfBuilderImpl) CDF(cdf float64) HttpCodeCdfBuilder {
	b.impl.Cdf = cdf
	return b
}

func (b *HttpCodeCdfBuilderImpl) DurationMean(mean float64) HttpCodeCdfBuilder {
	b.impl.DurationMean = mean
	return b
}

func (b *HttpCodeCdfBuilderImpl) DurationStddev(stddev float64) HttpCodeCdfBuilder {
	b.impl.DurationStddev = stddev
	return b
}

func (b *HttpCodeCdfBuilderImpl) AddHeader(name string, value string) HttpCodeCdfBuilder {
	b.impl.Headers.Add(name, value)
	return b
}

func (b *HttpCodeCdfBuilderImpl) Body(jsonBytes []byte) HttpCodeCdfBuilder {
	b.impl.JsonBody = jsonBytes
	return b
}

func (b *HttpCodeCdfBuilderImpl) Build() *HttpCodeCdf {
	defensiveCopy := b.impl
	return &defensiveCopy
}
