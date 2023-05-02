package server

import (
	"http-attenuator/data"
	"http-attenuator/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttpResponseBuilderDefaultCdf(t *testing.T) {
	builder := NewHttpCodeResponseBuilder()

	// Populate
	expectedCode := 999
	expectedWeight := 50
	expectedCdf := float64(1)
	expectedDuration, err := data.ParseDuration("10ms")
	assert.Nil(t, err)
	builder.Code(expectedCode)
	builder.Weight(expectedWeight)
	builder.Duration(expectedDuration)

	// build and verify
	resp := builder.Build()
	assert.NotNil(t, resp)
	assert.Equal(t, expectedCode, resp.Code)
	assert.Equal(t, expectedWeight, resp.Weight)
	assert.Equal(t, expectedCdf, resp.Cdf)
	assert.Equal(t, expectedDuration, resp.Duration)
}

func TestHttpResponseBuilderNonDefaultCdf(t *testing.T) {
	builder := NewHttpCodeResponseBuilder()

	// Populate
	expectedCode := 999
	expectedWeight := 50
	expectedCdf := float64(0.5)
	expectedDuration, err := data.ParseDuration("10ms")
	assert.Nil(t, err)
	builder.Code(expectedCode)
	builder.Weight(expectedWeight)
	builder.TotalWeight(100)
	builder.Duration(expectedDuration)

	// build and verify
	resp := builder.Build()
	assert.NotNil(t, resp)
	assert.Equal(t, expectedCode, resp.Code)
	assert.Equal(t, expectedWeight, resp.Weight)
	assert.Equal(t, expectedCdf, resp.Cdf)
	if !util.AlmostEqual(float64(expectedCdf), resp.Cdf) {
		t.Errorf("Expected %f, got %f", expectedCdf, resp.Cdf)
	}
}
