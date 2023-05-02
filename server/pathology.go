package server

import (
	"http-attenuator/data"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler interface {
	GetName() string
	Handle(c *gin.Context)
}

type Pathology interface {
	Handler
	data.HasCDF
}

type BasePathology struct {
	profile  string
	name     string
	weight   int
	cdf      float64
	duration data.HasDuration
}

type BasePathologyBuilder struct {
	impl BasePathology
}

func NewBasePathologyBuilder() *BasePathologyBuilder {
	return &BasePathologyBuilder{}
}

func (b *BasePathologyBuilder) Profile(profile string) *BasePathologyBuilder {
	b.impl.profile = profile
	return b
}

func (b *BasePathologyBuilder) Name(name string) *BasePathologyBuilder {
	b.impl.name = name
	return b
}

func (b *BasePathologyBuilder) Weight(weight int) *BasePathologyBuilder {
	b.impl.weight = weight
	return b
}

func (b *BasePathologyBuilder) CDF(cdf float64) *BasePathologyBuilder {
	b.impl.cdf = cdf
	return b
}

func (b *BasePathologyBuilder) DurationFunc(duration data.HasDuration) *BasePathologyBuilder {
	b.impl.duration = duration
	return b
}

func (b *BasePathologyBuilder) Build() (*BasePathology, error) {
	// Populate the CDF
	result := b.impl
	return &result, nil
}

func (f *BasePathology) GetName() string {
	return f.name
}

func (f *BasePathology) Handle(c *gin.Context) {
	// Default implementation is a NOP
}

// make BasePathology conform to the HasCDF interface
// so we can use our generic ChooseFromCDF() function
func (f *BasePathology) CDF() float64 {
	return f.cdf
}

func (f *BasePathology) SetCDF(cdf float64) {
	f.cdf = cdf
}

func (f *BasePathology) GetWeight() int {
	return f.weight
}

// BasePathology is-a HasDuration
func (f *BasePathology) GetDuration() *time.Duration {
	return nil
}

type PathologyDistribution struct {
	failureModes []data.HasCDF
}

func (pd *PathologyDistribution) Choose() Pathology {
	choice := data.ChooseFromCDF(rand.Float64(), pd.failureModes)
	if choice == nil {
		return nil
	}

	return choice.(Pathology)
}
