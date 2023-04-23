package broker

import (
	"fmt"
	"http-attenuator/data"
	config "http-attenuator/facade/config"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type ServiceMap interface {
	GetBackend(serviceName string, preference ...string) *Backend
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

		for serviceName, backendMap := range upstreamMap.(map[string]interface{}) {
			upstream := &Upstream{
				Name:     serviceName,
				Backends: make([]Backend, 0),
			}

			for label, backendConfig := range backendMap.(map[string]interface{}) {
				if label == "rule" {
					upstream.Rule = fmt.Sprint(backendConfig)
					continue
				}

				backend, err := NewBackendFromConfig(serviceName, label, backendConfig.(map[string]interface{}))
				if err != nil {
					panic(err)
				}
				upstream.Backends = append(upstream.Backends, *backend)
			}
			smi.upstream[strings.ToLower(upstream.Name)] = upstream
		}

		serviceMapInstance = smi
		rand.Seed(time.Now().UTC().UnixNano())
	})

	return serviceMapInstance
}

// GetBackend returns a backend which can handle a request
//
// TODO(john): different selection algorithms:
//
//   - random
//   - round-robin
//   - temperature-based
//   - priority
//   - lru
//   - mru
func (s *serviceMap) GetBackend(serviceName string, preference ...string) *Backend {
	if u, exists := s.upstream[strings.ToLower(serviceName)]; exists && u != nil {
		if len(u.Backends) > 0 {
			// Lame implementation - nested loops.  Yuck.
			if len(preference) > 0 {
				for _, preferredBackend := range preference {
					for _, backend := range u.Backends {
						if strings.EqualFold(preferredBackend, backend.Label) {
							return &backend
						}
					}
				}
			}

			// If we can't match our preference, or if we have not specified
			// a preference, then return a random one
			return &u.Backends[rand.Intn(len(u.Backends))]
		}
	}

	// No matching backend
	return nil
}
