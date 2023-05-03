package data

import (
	"strings"
)

var profileRegistry ProfileRegistry

func init() {
	profileRegistry = &ProfileRegistryImpl{
		pathologiesByName: make(map[string]PathologyProfile),
	}
}

func GetProfileRegistry() ProfileRegistry {
	return profileRegistry
}

type ProfileRegistry interface {
	Register(profile PathologyProfile)
	GetPathologyProfile(name string) PathologyProfile
}

type ProfileRegistryImpl struct {
	pathologiesByName map[string]PathologyProfile
}

func (hr *ProfileRegistryImpl) GetPathologyProfile(name string) PathologyProfile {
	return hr.pathologiesByName[strings.ToLower(name)]
}

func (hr *ProfileRegistryImpl) Register(profile PathologyProfile) {
	hr.pathologiesByName[strings.ToLower(profile.GetName())] = profile
}

// LoadRegistryFromConfig populates the profile registry singleton
// from the provided config
func LoadRegistryFromConfig(appConfig *AppConfig) error {
	// Instantiate the registry
	profileRegistry = &ProfileRegistryImpl{
		pathologiesByName: make(map[string]PathologyProfile),
	}

	// Parse and validate the config

	// Do any required backpatching
	//
	//	- populate 'name' where this is implied in the config
	//
	//	- calculate CDF values
	return nil
}
