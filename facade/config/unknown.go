package facade

import (
	"fmt"
)

type unknownConfigManager struct {
	typeName ConfigManagerType
}

func (cm *unknownConfigManager) Type() ConfigManagerType {
	return cm.typeName
}

func (cm *unknownConfigManager) GetString(name string, defaultValue ...string) (string, error) {
	return "", fmt.Errorf("ERROR|facade/config.unknown.GetString()| Unknown config type: %s", cm.typeName)
}

func (cm *unknownConfigManager) GetStringArray(name string, defaultValue ...[]string) ([]string, error) {
	return []string{}, fmt.Errorf("ERROR|facade/config.unknown.GetStringArray()| Unknown config type: %s", cm.typeName)
}

func (cm *unknownConfigManager) GetInt(name string, defaultValue ...int64) (int64, error) {
	return 0, fmt.Errorf("ERROR|facade/config.unknown.GetInt()| Unknown config type: %s", cm.typeName)
}

func (cm *unknownConfigManager) GetFloat(name string, defaultValue ...float64) (float64, error) {
	return 0, fmt.Errorf("ERROR|facade/config.unknown.GetInt()| Unknown config type: %s", cm.typeName)
}

func (cm *unknownConfigManager) GetBool(name string, defaultValue ...bool) (bool, error) {
	return false, fmt.Errorf("ERROR|facade/config.unknown.GetBool()| Unknown config type: %s", cm.typeName)
}

func (cm *unknownConfigManager) GetSecret(name string, defaultValue ...string) (string, error) {
	return "", fmt.Errorf("ERROR|facade/config.unknown.GetSecret()| Unknown config type: %s", cm.typeName)
}

func (cm *unknownConfigManager) GetValue(name string) (interface{}, error) {
	return nil, fmt.Errorf("ERROR|facade/config.unknown.GetValue()| Unknown config type: %s", cm.typeName)
}

func (cm *unknownConfigManager) SetString(name string, value string) error {
	return fmt.Errorf("ERROR|facade/config.unknown.SetString()| Unknown config type: %s", cm.typeName)
}

func (cm *unknownConfigManager) SetInt(name string, value int64) error {
	return fmt.Errorf("ERROR|facade/config.unknown.SetInt()| Unknown config type: %s", cm.typeName)
}

func (cm *unknownConfigManager) SetFloat(name string, value float64) error {
	return fmt.Errorf("ERROR|facade/config.unknown.SetInt()| Unknown config type: %s", cm.typeName)
}

func (cm *unknownConfigManager) SetBool(name string, value bool) error {
	return fmt.Errorf("ERROR|facade/config.unknown.SetBool()| Unknown config type: %s", cm.typeName)
}

func (cm *unknownConfigManager) Reset() {
}

func (cm *unknownConfigManager) GetAllValues(root string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("ERROR|facade/config.unknown.GetAllValues()| Unknown config type: %s", cm.typeName)
}
