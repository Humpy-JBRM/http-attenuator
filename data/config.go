package data

import (
	"fmt"
	"net/http"
	"os"
	"time"

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
	CONF_PATHOLOGY                = "config.pathology"
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
	for _, profile := range appConfig.Config.Profiles {
		if profile.HttpCode != nil {
			for code, response := range profile.HttpCode.Responses {
				// backpatch the http code
				response.Code = code

				// backpatch the duration
				// durationAsTime, err := parseDuration(response.Duration)
				// if err != nil {
				// 	return nil, fmt.Errorf("LoadConfig(%s): httpcode.%d: %s", configFile, code, err.Error())
				// }
				// response.durationAsTime = durationAsTime
			}
		}
	}

	return &appConfig, nil
}

type AppConfig struct {
	Config Config `yaml:"config"`
}

type Config struct {
	Profiles map[string]PathologyProfile `yaml:"pathologies"`
	Server   Server                      `yaml:"server"`
}

type PathologyProfile struct {
	HttpCode *HttpCodePathology `yaml:"httpcode"`
	Timeout  *TimeoutPathology  `yaml:"timeout"`
}

type HttpCodePathology struct {
	Weight    int                  `yaml:"weight"`
	Duration  string               `yaml:"duration"`
	Responses map[int]HttpResponse `yaml:"responses"`
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
	Pathology string `yaml:"pathology"`
}
