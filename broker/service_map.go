package broker

import (
	"fmt"
	"http-attenuator/data"
	config "http-attenuator/facade/config"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
)

type ServiceMap interface {
	GetBackend(serviceName string, preference ...string) *Backend
}

type serviceMap struct {
	upstream map[string]*BrokeredServiceImpl
}

var serviceMapInstance ServiceMap
var serviceMapOnce sync.Once

func GetServiceMap() ServiceMap {
	serviceMapOnce.Do(func() {
		smi := &serviceMap{
			upstream: make(map[string]*BrokeredServiceImpl),
		}

		upstreamMap, err := config.Config().GetValue(data.CONF_BROKER_UPSTREAM)
		if err != nil {
			panic(err)
		}

		totalWeight := float64(0)
		for serviceName, backendMap := range upstreamMap.(map[string]interface{}) {
			upstream := &BrokeredServiceImpl{
				Name:     serviceName,
				Backends: make([]Backend, 0),
			}

			for label, backendConfig := range backendMap.(map[string]interface{}) {
				if label == "cost" {
					// TODO(john): TRIBE integration
					continue
				}
				if label == "rule" {
					upstream.Rule = fmt.Sprint(backendConfig)
					continue
				}

				backend, err := NewBackendFromConfig(serviceName, label, backendConfig.(map[string]interface{}))
				if err != nil {
					panic(err)
				}
				totalWeight += backend.Weight
				upstream.Backends = append(upstream.Backends, *backend)
			}

			// Deal with any weightings in the config.
			//
			// Put the backends in a CDF so we can select according to the
			// weighting.
			cdf := float64(0)
			for i := 0; i < len(upstream.Backends); i++ {
				upstream.Backends[i].ProbabilityCDF = cdf + upstream.Backends[i].Weight/totalWeight
				cdf += upstream.Backends[i].Weight / totalWeight
			}
			smi.upstream[strings.ToLower(upstream.Name)] = upstream
		}

		// Sort the upstream backends according to their cumulative
		// probability.
		//
		// This makes it easier to choose a backend when the rule is
		// 'weighted'
		for _, upstream := range smi.upstream {
			// sort.SliceStable() preserves ordering of equal elements
			sort.SliceStable(upstream.Backends, func(i, j int) bool {
				return upstream.Backends[i].ProbabilityCDF < upstream.Backends[j].ProbabilityCDF
			})
		}

		// Add the default gateway / forward proxy
		defaultGateway := &BrokeredServiceImpl{
			Name: "gateway",
			Backends: []Backend{
				*NewDefaultGateway(),
			},
		}
		smi.upstream["gateway"] = defaultGateway
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
			switch u.Rule {
			case "weighted":
				// We choose according to the weight specified in the config file
				//
				// Choose a number: 0.0 < number <= 1.0
				//
				// Find the backed in the CDF covering this value
				probability := rand.Float64()
				for _, backend := range u.Backends {
					if backend.ProbabilityCDF >= 1.0 || backend.ProbabilityCDF > probability {
						return &backend
					}
				}

				// For some reason, we could not get a backend from the CDF.
				//
				// Default to random
				fallthrough

			case "random":
				fallthrough
			default:
				return &u.Backends[rand.Intn(len(u.Backends))]
			}
		}
	}

	// No matching backend
	return nil
}
