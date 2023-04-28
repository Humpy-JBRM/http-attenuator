package server

import (
	"http-attenuator/data"
	"strings"
)

var handlerRegistry HandlerRegistry

func init() {
	handlerRegistry = &HandlerRegistryImpl{
		handlersByName: make(map[string]data.Handler),
	}
}

func GetRegistry() HandlerRegistry {
	return handlerRegistry
}

type HandlerRegistry interface {
	Register(handler data.Handler) data.Handler
	GetHandler(name string) data.Handler
}

type HandlerRegistryImpl struct {
	handlersByName map[string]data.Handler
}

func (hr *HandlerRegistryImpl) GetHandler(name string) data.Handler {
	return hr.handlersByName[strings.ToLower(name)]
}

func (hr *HandlerRegistryImpl) Register(handler data.Handler) data.Handler {
	hr.handlersByName[strings.ToLower(handler.GetName())] = handler
	return handler
}
