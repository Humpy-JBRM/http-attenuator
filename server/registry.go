package server

import (
	"http-attenuator/data"
	"strings"
)

var handlerRegistry HandlerRegistry

func init() {
	handlerRegistry = newRegistry()
}

func newRegistry() HandlerRegistry {
	return &HandlerRegistryImpl{
		handlersByName: make(map[string]Handler),
	}
}

func GetRegistry() HandlerRegistry {
	return handlerRegistry
}

type HandlerRegistry interface {
	Register(handler Handler) Handler
	GetHandler(name string) Handler
}

type HandlerRegistryImpl struct {
	handlersByName map[string]Handler
}

func (hr *HandlerRegistryImpl) GetHandler(name string) Handler {
	return hr.handlersByName[strings.ToLower(name)]
}

func (hr *HandlerRegistryImpl) Register(handler Handler) Handler {
	hr.handlersByName[strings.ToLower(handler.GetName())] = handler
	return handler
}

// LoadRegistryFromConfig populates the pathology registry singleton
// from the provided config
func LoadRegistryFromConfig(appConfig *data.AppConfig) error {
	// Instantiate the registry
	handlerRegistry = &HandlerRegistryImpl{
		handlersByName: make(map[string]Handler),
	}

	// Parse and validate the config

	// Do any required backpatching
	//
	//	- populate 'name' where this is implied in the config
	//
	//	- calculate CDF values
	return nil
}
