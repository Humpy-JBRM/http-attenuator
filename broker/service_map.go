package broker

import (
	"fmt"
	"http-attenuator/data"
	config "http-attenuator/facade/config"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Backend struct {
	Label string   `json:"label"`
	Url   *url.URL `json:"url"`
}

func (b *Backend) String() string {
	return fmt.Sprintf("%s@%s", b.Label, b.Url.String())
}

type Upstream struct {
	Name     string    `json:"name"`
	Rule     string    `json:"rule"`
	Backends []Backend `json:"backends"`
}

type ServiceMap interface {
	GetBackend(serviceName string) *Backend
}

type serviceMap struct {
	upstream map[string]*Upstream
}

var serviceMapInstance ServiceMap
var serviceMapOnce sync.Once

func GetServiceMap() ServiceMap {
	serviceMapOnce.Do(func() {
		smi := &serviceMap{
			upstream: make(map[string]*Upstream),
		}

		upstreamMap, err := config.Config().GetValue(data.CONF_BROKER_UPSTREAM)
		if err != nil {
			panic(err)
		}

		for name, backendMap := range upstreamMap.(map[string]interface{}) {
			upstream := &Upstream{
				Name:     name,
				Backends: make([]Backend, 0),
			}

			for label, endpoint := range backendMap.(map[string]interface{}) {
				if label == "rule" {
					upstream.Rule = fmt.Sprint(endpoint)
					continue
				}
				url, err := url.Parse(fmt.Sprint(endpoint))
				if err != nil {
					panic(err)
				}
				upstream.Backends = append(
					upstream.Backends,
					Backend{
						Label: label,
						Url:   url,
					},
				)
			}
			smi.upstream[strings.ToLower(upstream.Name)] = upstream
		}

		serviceMapInstance = smi
		rand.Seed(time.Now().UnixNano())
	})

	return serviceMapInstance
}

func (s *serviceMap) GetBackend(serviceName string) *Backend {
	if u, exists := s.upstream[strings.ToLower(serviceName)]; exists && u != nil {
		if len(u.Backends) > 0 {
			return &u.Backends[rand.Intn(len(u.Backends))]
		}
	}

	// No matching backend
	return nil
}
