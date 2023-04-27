package server

import "strings"

var handlerRegistry HandlerRegistry

func init() {
	handlerRegistry = &HandlerRegistryImpl{
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
