package data

import (
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
)

type Broker struct {
	// Name   string `yaml:"name" json:"name"`
	Listen string `yaml:"listen" json:"listen"`
	// Enable bool   `yaml:"enable" json:"enable"`

	// // Mapping of host header value -> implementation
	// Hosts map[string]*ServerHost
	UpstreamFromConfig map[string]*UpstreamImpl `yaml:"upstream" json:"upstream"`

	// These are backpatched
	upstream map[string]Upstream
}

func (b *Broker) backpatch() error {
	b.upstream = make(map[string]Upstream)
	for upstreamServiceName, upstreamService := range b.UpstreamFromConfig {
		upstreamService.serviceName = upstreamServiceName
		err := upstreamService.backpatch()
		if err != nil {
			return err
		}
		b.upstream[upstreamServiceName] = upstreamService
	}

	return nil
}

type Upstream interface {
	Handler
}

type UpstreamImpl struct {
	// this is backpatched
	serviceName string
	Cost        CostFromConfig                  `yaml:"cost" json:"cost"`
	Backends    map[string]*UpstreamBackendImpl `yaml:"backends" json:"backends"`
	Rule        string                          `yaml:"rule" json:"rule"`
	Pathology   string                          `yaml:"pathology" json:"pathology"`
	// These are backpatched
	backendCDF []HasCDF
	cost       Cost
	rng        *rand.Rand
}

func (u *UpstreamImpl) backpatch() error {
	u.rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	// Backpatch the cost
	u.cost = &CostImpl{
		coins: u.Cost,
	}

	// Backpatch the backends CDF
	totalWeight := 0
	for backendName, upstreamBackend := range u.Backends {
		upstreamBackend.backendName = backendName
		totalWeight += upstreamBackend.Weight
	}
	u.backendCDF = make([]HasCDF, 0)
	for _, upstreamBackend := range u.Backends {
		upstreamBackend.cdf = float64(upstreamBackend.Weight) / float64(totalWeight)
		u.backendCDF = append(u.backendCDF, upstreamBackend)
	}

	return nil
}

func (u *UpstreamImpl) GetName() string {
	return u.serviceName
}

func (u *UpstreamImpl) Handle(c *gin.Context) {
}

func (u *UpstreamImpl) ChooseBackend() Upstream {
	backend := Choose(u.Rule, u.backendCDF, u.rng)
	if backend == nil {
		return nil
	}

	return backend.(Upstream)
}

func (u *UpstreamImpl) backpatchCdf() {
	u.backendCDF = make([]HasCDF, 0)
}

type UpstreamBackend interface {
	HasCDF
	GetName() string
	GetCost() Cost
}

type UpstreamBackendImplFromConfig map[string]*UpstreamBackendImpl

type UpstreamBackendImpl struct {
	// this is backpatched
	backendName string
	Impl        string `yaml:"impl" json:"impl"`
	Url         string `yaml:"url" json:"url"`
	Weight      int    `yaml:"weight" json:"weight"`
	Pathology   string `yaml:"pathology" json:"pathology"`

	// This is used to override the default cost for this upstream.
	//
	// It allows us to (for instance) implement some kind of
	// ChooseCheapest()
	Cost Cost `yaml:"cost,omitempty" json:"impl,omitempty"`

	// This is backpatched
	cdf float64
}

func (u *UpstreamBackendImpl) GetName() string {
	return u.backendName
}

func (u *UpstreamBackendImpl) CDF() float64 {
	return u.cdf
}

func (u *UpstreamBackendImpl) SetCDF(cdf float64) {
	u.cdf = cdf
}

func (u *UpstreamBackendImpl) GetWeight() int {
	return u.Weight
}

func (u *UpstreamBackendImpl) GetCost() Cost {
	return u.Cost
}

// type UpstreamBackendBuilder struct {
// 	impl UpstreamBackendImpl
// }

// func NewUpstreamBackendBuilder() *UpstreamBackendBuilder {
// 	return &UpstreamBackendBuilder{}
// }

// func (ub *UpstreamBackendBuilder) FromConfig(ubFromConfig UpstreamFromConfig) *UpstreamBackendBuilder {
// 	ub.SetImpl(ubFromConfig["impl"])
// 	return ub
// }

// func (ub *UpstreamBackendBuilder) SetImpl(impl string) *UpstreamBackendBuilder {
// 	ub.impl.Impl = impl
// 	return ub
// }

// func (ub *UpstreamBackendBuilder) SetUrl(url string) *UpstreamBackendBuilder {
// 	ub.impl.Url = url
// 	return ub
// }

// func (ub *UpstreamBackendBuilder) SetWeight(weight int) *UpstreamBackendBuilder {
// 	ub.impl.Weight = weight
// 	return ub
// }

// func (ub *UpstreamBackendBuilder) SetCost(cost Cost) *UpstreamBackendBuilder {
// 	ub.impl.Cost = cost
// 	return ub
// }

// func (ub *UpstreamBackendBuilder) Build() UpstreamBackend {
// 	panic("TODO(john): bckpatch CDF")
// 	return &ub.impl
// }
