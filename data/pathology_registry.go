package data

import (
	"strings"
)

var pathologyRegistry PathologyRegistry

func init() {
	pathologyRegistry = newRegistry()
}

func newRegistry() PathologyRegistry {
	return &PathologyRegistryImpl{
		pathologiesByName: make(map[string]Pathology),
	}
}

func GetRegistry() PathologyRegistry {
	return pathologyRegistry
}

type PathologyRegistry interface {
	Register(pathology Pathology) Pathology
	GetPathology(name string) Pathology
}

type PathologyRegistryImpl struct {
	pathologiesByName map[string]Pathology
}

func (hr *PathologyRegistryImpl) GetPathology(name string) Pathology {
	return hr.pathologiesByName[strings.ToLower(name)]
}

func (hr *PathologyRegistryImpl) Register(pathology Pathology) Pathology {
	hr.pathologiesByName[strings.ToLower(pathology.GetName())] = pathology
	return pathology
}

// LoadRegistryFromConfig populates the pathology registry singleton
// from the provided config
func LoadRegistryFromConfig(appConfig *AppConfig) error {
	// Instantiate the registry
	pathologyRegistry = &PathologyRegistryImpl{
		pathologiesByName: make(map[string]Pathology),
	}

	// Parse and validate the config

	// Do any required backpatching
	//
	//	- populate 'name' where this is implied in the config
	//
	//	- calculate CDF values
	return nil
}
