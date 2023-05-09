package data

import (
	"log"
	"strings"
)

var registry AttenuatorRegistry

func init() {
	registry = &attenuatorRegistryImpl{
		attenuators: make(map[string]Attenuator),
	}
}

func GetAttenuatorRegistry() AttenuatorRegistry {
	return registry
}

type AttenuatorRegistry interface {
	GetAttenuator(name string) Attenuator
	RegisterAttenuator(name string, attenuator Attenuator)
}

type attenuatorRegistryImpl struct {
	attenuators map[string]Attenuator
}

func (ar *attenuatorRegistryImpl) GetAttenuator(name string) Attenuator {
	return ar.attenuators[strings.ToLower(strings.TrimSpace(name))]
}

func (ar *attenuatorRegistryImpl) RegisterAttenuator(name string, attenuator Attenuator) {
	ar.attenuators[strings.ToLower(strings.TrimSpace(name))] = attenuator
	log.Printf("INFO|RegisterAttenuator()|Registering attenuator '%s'|", name)
}
