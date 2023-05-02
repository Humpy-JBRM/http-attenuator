package server

import (
	"http-attenuator/data"
	"net/http"
	"time"
)

type HttpPayload struct {
	Headers  http.Header
	JsonBody []byte
}

type HttpCodeResponse struct {
	HttpPayload
	Code     int
	Weight   int
	Cdf      float64
	Duration data.HasDuration
}

type HttpCodeResponseBuilder struct {
	impl        HttpCodeResponse
	totalWeight int
}

func NewHttpCodeResponseBuilder() *HttpCodeResponseBuilder {
	return &HttpCodeResponseBuilder{}
}

func (b *HttpCodeResponseBuilder) Code(code int) *HttpCodeResponseBuilder {
	b.impl.Code = code
	return b
}

func (b *HttpCodeResponseBuilder) Weight(weight int) *HttpCodeResponseBuilder {
	b.impl.Weight = weight
	return b
}

// TotalWeight sets the total weight so the cdf can be backpatched
func (b *HttpCodeResponseBuilder) TotalWeight(totalWeight int) *HttpCodeResponseBuilder {
	b.totalWeight = totalWeight
	return b
}

func (b *HttpCodeResponseBuilder) AddHeader(name string, values ...string) *HttpCodeResponseBuilder {
	for _, v := range values {
		b.impl.Headers.Add(name, v)
	}
	return b
}

func (b *HttpCodeResponseBuilder) Body(body []byte) *HttpCodeResponseBuilder {
	b.impl.JsonBody = body
	return b
}

func (b *HttpCodeResponseBuilder) Duration(duration data.HasDuration) *HttpCodeResponseBuilder {
	b.impl.Duration = duration
	return b
}

func (b *HttpCodeResponseBuilder) Build() *HttpCodeResponse {
	result := b.impl

	// backpatch the cdf
	if b.totalWeight == 0 {
		b.totalWeight = result.Weight
	}
	result.Cdf = float64(result.Weight) / float64(b.totalWeight)
	return &result
}

// make returnCodeCdf conform to the HasCDF interface
// so we can use our generic ChooseFromCDF() function
func (c *HttpCodeResponse) CDF() float64 {
	return c.Cdf
}

func (c *HttpCodeResponse) SetCDF(cdf float64) {
	c.Cdf = cdf
}

func (c *HttpCodeResponse) GetWeight() int {
	return c.Weight
}

// GetDuration calculates a sleep time based on the
// distribution
//
// This is to conform to the HasDuration interface
func (c *HttpCodeResponse) GetDuration() *time.Duration {
	if c.Duration == nil {
		return nil
	}

	return c.Duration.GetDuration()
	// if c.DurationMean > 0 && c.DurationStddev > 0 {
	// 	durationSeconds := rand.NormFloat64()*c.DurationStddev + c.DurationMean
	// 	if durationSeconds > 0 {
	// 		duration := time.Duration(int64(durationSeconds*1000)) * time.Millisecond
	// 		return &duration
	// 		//pathologyLatency.WithLabelValues(fmt.Sprint(pathology), h.name, c.Request.Method).Add(sleepSeconds * 1000)
	// 	}

	// 	return nil
	// }

	// if c.DurationMillis > 0 {
	// 	duration := time.Duration(c.DurationMillis) * time.Millisecond
	// 	return &duration
	// }

	// return nil
}
