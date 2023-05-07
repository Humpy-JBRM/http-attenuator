package server

import "http-attenuator/data"

type ServerBuilder interface {
	FromConfig(appConfig *data.AppConfig) (ServerBuilder, error)
	Build() (*data.Server, error)
}

type ServerBuilderImpl struct {
	impl data.Server
}

func NewServerBuilder() ServerBuilder {
	return &ServerBuilderImpl{}
}

// FromConfig populates the pathology registry
func (b *ServerBuilderImpl) FromConfig(appConfig *data.AppConfig) (ServerBuilder, error) {
	b.impl = appConfig.Config.Server

	// Validate the config
	for _, u := range appConfig.Config.Broker.UpstreamFromConfig {
		err := u.Backpatch()
		if err != nil {
			return b, err
		}
	}

	// Do any required backpatching
	return b, nil
}

func (b *ServerBuilderImpl) Build() (*data.Server, error) {
	return &b.impl, nil
}
