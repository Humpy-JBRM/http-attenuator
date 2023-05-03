package data

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	CONF_ATTENUATOR_LISTEN        = "config.attenuator.listen"
	CONF_ATTENUATOR_QUEUESIZE     = "config.attenuator.queue_size"
	CONF_BILLING_ENABLE           = "config.billing.enable"
	CONF_BROKER_LISTEN            = "config.broker.listen"
	CONF_BROKER_UPSTREAM          = "config.broker.upstream"
	CONF_CONFIG_LISTEN            = "config.config.listen"
	CONF_GATEWAY_LISTEN           = "config.gateway.listen"
	CONF_GATEWAY_RECORD           = "config.gateway.record"
	CONF_GATEWAY_RECORD_REQUESTS  = "config.gateway.record.requests"
	CONF_GATEWAY_RECORD_RESPONSES = "config.gateway.record.responses"
	CONF_PATHOLOGIES              = "config.pathologies"
	CONF_PROXY_ENABLE             = "config.proxy.enable"
	CONF_PROXY_LISTEN             = "config.proxy.listen"
	CONF_REDIS_HOST               = "config.redis.host"
	CONF_REDIS_POOLSIZE           = "config.redis.pool_size"
	CONF_REDIS_TIMEOUT            = "config.redis.timeout"
	CONF_SERVER_ENABLE            = "config.server.enable"
	CONF_SERVER_LISTEN            = "config.server.listen"
)

func LoadConfig(configFile string) (*AppConfig, error) {
	configBytes, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("LoadCOnfig(%s): %s", configFile, err.Error())
	}

	appConfig := AppConfig{
		Config: Config{
			pathologyProfiles: make(map[string]PathologyProfile),
		},
	}
	err = yaml.Unmarshal(configBytes, &appConfig)
	if err != nil {
		return nil, fmt.Errorf("LoadConfig(%s): %s", configFile, err.Error())
	}

	for profileName, profile := range appConfig.Config.PathologiesFromConfig {
		profileInstance := &PathologyProfileImpl{
			PathologyProfileFromConfig: profile,
			pathologyCdf:               make([]HasCDF, 0),
			rng:                        rand.New(rand.NewSource(time.Now().UnixNano())),
		}
		appConfig.Config.pathologyProfiles[profileName] = profileInstance
		totalProfileWeight := 0
		for name, pathology := range profile {
			totalProfileWeight += pathology.Weight

			// backpatch the name and the profile this pathology belongs to
			pathology.name = name
			pathology.profile = profileName
			totalWeight := 0
			for code, response := range pathology.Responses {
				totalWeight += response.Weight

				// backpatch the http code
				response.Code = code

				// backpatch the duration
				duration := "0ms"
				if response.Duration == "" {
					// Inherit from parent
					if pathology.Duration != "" {
						duration = pathology.Duration
					}
				} else {
					duration = response.Duration
				}
				durationAsTime, err := ParseDuration(duration)
				if err != nil {
					return nil, fmt.Errorf("LoadConfig(%s): httpcode.%d: %s", configFile, code, err.Error())
				}
				response.durationConfig = durationAsTime
			}

			// Backpatch the cdf for the various responses
			for _, response := range pathology.Responses {
				if len(pathology.Responses) == 1 {
					response.cdf = float64(1)
					continue
				}

				response.cdf = float64(response.Weight) / float64(totalWeight)
			}
		}
		// Backpatch the cdf for the various pathologies in the profile
		for _, profile := range appConfig.Config.PathologiesFromConfig {
			for _, pathology := range profile {
				pathology.SetCDF(float64(pathology.Weight) / float64(totalProfileWeight))
			}
		}
	}

	// Make sure the pathology profiles are registered
	for name, profile := range appConfig.Config.pathologyProfiles {
		// backpatch the name
		profile.(*PathologyProfileImpl).name = name
		GetProfileRegistry().Register(profile)

		// Backpatch the CDF for the pathologies in the profiles
		totalWeight := 0
		for _, pathology := range profile.(*PathologyProfileImpl).PathologyProfileFromConfig {
			totalWeight += pathology.Weight
		}

		for _, pathology := range profile.(*PathologyProfileImpl).PathologyProfileFromConfig {
			pathology.cdf = float64(pathology.Weight) / float64(totalWeight)
			profile.(*PathologyProfileImpl).pathologyCdf = append(profile.(*PathologyProfileImpl).pathologyCdf, pathology)
		}

		sort.Slice(profile.(*PathologyProfileImpl).pathologyCdf, func(i, j int) bool {
			return profile.(*PathologyProfileImpl).pathologyCdf[i].CDF() > profile.(*PathologyProfileImpl).pathologyCdf[j].CDF()
		})
	}

	// Backpatch the servers with the actual pathology profile instance
	// to be used
	for _, serverHost := range appConfig.Config.Server.Hosts {
		serverHost.pathologyProfile = GetProfileRegistry().GetPathologyProfile(serverHost.PathologyProfileName)
	}
	return &appConfig, nil
}

type AppConfig struct {
	Config Config `yaml:"config" json:"config"`
}

type Config struct {
	PathologiesFromConfig map[string]PathologyProfileFromConfig `yaml:"pathologies" json:"pathologies"`
	Server                Server                                `yaml:"server" json:"server"`

	// These are backpatched
	pathologyProfiles map[string]PathologyProfile
}

func (c *Config) GetProfile(name string) PathologyProfile {
	return c.pathologyProfiles[name]
}
