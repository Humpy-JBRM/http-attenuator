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

// FromConfig populates
func (b *ServerBuilderImpl) FromConfig(appConfig *data.AppConfig) (ServerBuilder, error) {
	b.impl = appConfig.Config.Server

	// Validate the config

	// Do any required backpatching
	return b, nil
}

func (b *ServerBuilderImpl) Build() (*data.Server, error) {
	panic("IMPLEMENT ME")
}
