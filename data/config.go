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

	var appConfig AppConfig
	err = yaml.Unmarshal(configBytes, &appConfig)
	if err != nil {
		return nil, fmt.Errorf("LoadConfig(%s): %s", configFile, err.Error())
	}

	// If we have a httpcode pathology in any of the profiles, then we need
	// to backpatch the code.
	//
	// This is because we reuse the HttpCode in other places and repeating
	// the 'code: NNN' is redundant when the code is the key to a map.
	for profileName, profile := range appConfig.Config.Pathologies {
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
		for _, profile := range appConfig.Config.Pathologies {
			for _, pathology := range profile {
				pathology.SetCDF(float64(pathology.Weight) / float64(totalProfileWeight))
			}
		}
	}

	return &appConfig, nil
}

type AppConfig struct {
	Config Config `yaml:"config"`
}

type Config struct {
	Pathologies map[string]PathologyProfile `yaml:"pathologies"`
	Server      Server                      `yaml:"server"`
}

type PathologyProfile map[string]*PathologyImpl

func (pp PathologyProfile) GetPathology(name string) Pathology {
	return pp[name]
}

type PathologyImpl struct {
	Weight    int                   `yaml:"weight"`
	Duration  string                `yaml:"duration"`
	Responses map[int]*HttpResponse `yaml:"responses"`

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

func (p *PathologyImpl) GetProfile() string {
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
	Code     int         `yaml:"code"`
	Weight   int         `yaml:"weight"`
	Duration string      `yaml:"duration"`
	Headers  http.Header `yaml:"headers"`
	Body     string      `yaml:"body"`

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
	Millis   int64         `yaml:"millis"`
	Weight   int           `yaml:"weight"`
	Response *HttpResponse `yaml:"response"`
}

type Server struct {
	Name   string `yaml:"name"`
	Listen string `yaml:"listen"`
	Enable bool   `yaml:"enable"`

	// Mapping of host header value -> implementation
	Hosts map[string]ServerHost
}

type ServerHost struct {
	PathologyName string `yaml:"pathology"`
}
