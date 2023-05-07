package data

import "github.com/gin-gonic/gin"

type Broker interface {
	Handler
	GetUpstream(serviceName string) Upstream
}

type BrokerImpl struct {
	Name   string `yaml:"name" json:"name"`
	Listen string `yaml:"listen" json:"listen"`
	// Enable bool   `yaml:"enable" json:"enable"`

	// // Mapping of host header value -> implementation
	// Hosts map[string]*ServerHost
	UpstreamFromConfig map[string]*UpstreamImpl `yaml:"upstream" json:"upstream"`

	// These are backpatched
	upstream map[string]Upstream
}

func (b *BrokerImpl) GetName() string {
	return b.Name
}

func (b *BrokerImpl) Handle(c *gin.Context) {
	panic("No default implmentation for BrokerImpl.Handle()")
}

func (b *BrokerImpl) GetUpstream(serviceName string) Upstream {
	return b.upstream[serviceName]
}

func (b *BrokerImpl) Backpatch() error {
	b.upstream = make(map[string]Upstream)
	for upstreamServiceName, upstreamService := range b.UpstreamFromConfig {
		upstreamService.serviceName = upstreamServiceName
		err := upstreamService.Backpatch()
		if err != nil {
			return err
		}
		b.upstream[upstreamServiceName] = upstreamService
	}

	return nil
}
