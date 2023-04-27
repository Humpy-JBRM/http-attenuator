package facade

// ConfigManager encapsulates the behaviour that we want from a
// configuration manager component, abstracting away the
// implementation.
//
// The naiive implementation runs entirely in memory, and has no
// external dependencies.
//
// A production implementation would be a facade for vault / k8s secrets.

type ConfigApi interface {
	GetString(string, ...string) (string, error)
	GetStringArray(string, ...[]string) ([]string, error)
	GetInt(string, ...int64) (int64, error)
	GetFloat(string, ...float64) (float64, error)
	GetBool(string, ...bool) (bool, error)
	GetSecret(string, ...string) (string, error)
	GetValue(string) (interface{}, error)
	SetString(string, string) error
	SetInt(string, int64) error
	SetFloat(string, float64) error
	SetBool(string, bool) error
	GetAllValues(root string) (map[string]interface{}, error)
}

type ConfigManager interface {
	ConfigApi
	Type() ConfigManagerType
	Reset()
}

// ConfigManagerFactory produces ConfigManager instances based on
// the criteria we set.
//
// This is how we can flip implementations without refactoring
// any code
type ConfigManagerFactory interface {
	SetType(cmType ConfigManagerType) ConfigManagerFactory
	New() (ConfigManager, error)
}
