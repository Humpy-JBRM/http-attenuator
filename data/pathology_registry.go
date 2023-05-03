package data

import (
	"strings"
)

var profileRegistry ProfileRegistry

func init() {
	profileRegistry = &ProfileRegistryImpl{
		pathologiesByName: make(map[string]*PathologyProfileImpl),
	}
}

func GetProfileRegistry() ProfileRegistry {
	return profileRegistry
}

type ProfileRegistry interface {
	Register(profile *PathologyProfileImpl)
	GetPathologyProfile(name string) *PathologyProfileImpl
}

type ProfileRegistryImpl struct {
	pathologiesByName map[string]*PathologyProfileImpl
}

func (hr *ProfileRegistryImpl) GetPathologyProfile(name string) *PathologyProfileImpl {
	return hr.pathologiesByName[strings.ToLower(name)]
}

func (hr *ProfileRegistryImpl) Register(profile *PathologyProfileImpl) {
	hr.pathologiesByName[strings.ToLower(profile.GetName())] = profile
}

// LoadRegistryFromConfig populates the profile registry singleton
// from the provided config
func LoadRegistryFromConfig(appConfig *AppConfig) error {
	// Instantiate the registry
	profileRegistry = &ProfileRegistryImpl{
		pathologiesByName: make(map[string]*PathologyProfileImpl),
	}

	// Parse and validate the config

	// Do any required backpatching
	//
	//	- populate 'name' where this is implied in the config
	//
	//	- calculate CDF values
	return nil
}
