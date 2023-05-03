package data

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

const (
	CONF_ATTENUATOR_LISTEN        = "config.attenuator.listen"
	CONF_ATTENUATOR_QUEUESIZE     = "config.attenuator.queue_size"
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
	configBytes, err := os.ReadFile(os.Getenv("CONFIG_FILE"))
	if err != nil {
		return nil, fmt.Errorf("LoadCOnfig(%s): %s", configFile, err.Error())
	}

	appConfig := AppConfig{
		Config: Config{
			pathologyProfiles: make(map[string]*PathologyProfileImpl),
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
		profile.name = name
		GetProfileRegistry().Register(profile)
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
	pathologyProfiles map[string]*PathologyProfileImpl
}

func (c *Config) GetProfile(name string) *PathologyProfileImpl {
	return c.pathologyProfiles[name]
}

type PathologyProfileFromConfig map[string]*PathologyImpl

type PathologyProfile interface {
	Handler
}

type PathologyProfileImpl struct {
	name string
	PathologyProfileFromConfig

	// The pathologies in this profile as a CDF
	pathologyCdf []HasCDF
}

func (pp *PathologyProfileImpl) GetName() string {
	return pp.name
}

func (pp *PathologyProfileImpl) GetPathology(name string) Pathology {
	return pp.PathologyProfileFromConfig[name]
}

type PathologyImpl struct {
	Weight    int                   `yaml:"weight" json:"weight"`
	Duration  string                `yaml:"duration" json:"duration"`
	Responses map[int]*HttpResponse `yaml:"responses" json:"responses"`

	// The CDF when this pathology is part of a profile
	cdf float64

	// These get backpatched in LoadConfig()
	name              string
	profile           string
	rng               *rand.Rand
	responsesAsHasCDF []HasCDF
}

func (p *PathologyImpl) GetName() string {
	return p.name
}

func (p *PathologyImpl) GetProfileName() string {
	return p.profile
}

// PathologyImpl must conform to Pathology (HasCDF) duck-type
func (p *PathologyImpl) CDF() float64 {
	return p.cdf
}

func (p *PathologyImpl) SetCDF(cdf float64) {
	p.cdf = cdf
}

func (p *PathologyImpl) GetWeight() int {
	return p.Weight
}

// SelectResponse selects the HttpResponse to be returned
// based on the cdf
func (p *PathologyImpl) SelectResponse() *HttpResponse {
	if len(p.Responses) == 0 {
		return nil
	}
	for _, resp := range p.Responses {
		return resp
	}

	if p.responsesAsHasCDF == nil {
		return nil
	}
	return ChooseFromCDF(p.rng.Float64(), p.responsesAsHasCDF).(*HttpResponse)
}

// Satisfy the Handler duck type
func (p *PathologyImpl) Handle(c *gin.Context) {
	pathologyRequests.WithLabelValues(p.profile, p.name, c.Request.Method).Inc()
	resp := p.SelectResponse()
	if resp == nil {
		log.Printf("%s.Handle(%s): no response configured", p.name, c.Request.URL.String())
		return
	}

	// Response code
	c.Status(resp.Code)

	// Headers
	for headerName, values := range resp.Headers {
		for _, value := range values {
			c.Writer.Header().Add(headerName, value)
		}
	}

	// Response body
	c.Writer.Write([]byte(resp.Body))
}

type HttpResponse struct {
	// This needs to be backpatched in the case of a httpcode,
	// because it lives in a map
	Code     int         `yaml:"code" json:"code"`
	Weight   int         `yaml:"weight" json:"weight"`
	Duration string      `yaml:"duration" json:"duration"`
	Headers  http.Header `yaml:"headers" json:"headers"`
	Body     string      `yaml:"body" json:"body"`

	// this needs to be backpatched because it is derived
	// from the config value (which could be a formula)
	durationConfig HasDuration

	// this needs to be backpatched so we can select responses
	// according to a cdf
	cdf float64
}

// HttpResponse should satisfy the HasDuration interface
func (r *HttpResponse) GetDuration() *time.Duration {
	if r.durationConfig == nil {
		return nil
	}

	// delegate to the duration config to give us a duration
	// according to its distribution
	return r.durationConfig.GetDuration()
}

// HttpResponse should satisfy the HasCDF interface
func (r *HttpResponse) CDF() float64 {
	return r.cdf
}

// HttpResponse should satisfy the HasCDF interface
func (r *HttpResponse) SetCDF(cdf float64) {
	r.cdf = cdf
}

// HttpResponse should satisfy the HasCDF interface
func (r *HttpResponse) GetWeight() int {
	return r.Weight
}

type TimeoutPathology struct {
	Millis   int64         `yaml:"millis" json:"millis"`
	Weight   int           `yaml:"weight" json:"weight"`
	Response *HttpResponse `yaml:"response" json:"response"`
}

type Server struct {
	Name   string `yaml:"name" json:"name"`
	Listen string `yaml:"listen" json:"listen"`
	Enable bool   `yaml:"enable" json:"enable"`

	// Mapping of host header value -> implementation
	Hosts map[string]*ServerHost
}

type ServerHost struct {
	PathologyProfileName string `yaml:"pathology" json:"pathology"`

	// The actual pathology profile instance gets backpatched
	pathologyProfile *PathologyProfileImpl
}

func (s *ServerHost) GetPathologyProfile() *PathologyProfileImpl {
	return s.pathologyProfile
}
