package facade

import (
	"testing"
)

func TestSetGetString(t *testing.T) {
	name := "string.variable"
	value := "SomeValue"
	configManager = nil
	Config().Reset()

	testSetGetString(t, Config(), name, value)
}

func TestGetStringDefaultValue(t *testing.T) {
	name := "string.variable"
	value := "defaultValue"
	configManager = nil
	Config().Reset()

	testGetStringDefault(t, Config(), name, value)

}

func TestSetGetInt(t *testing.T) {
	name := "int.variable"
	value := int64(12345)
	configManager = nil
	Config().Reset()

	testSetGetInt(t, Config(), name, value)
}

func TestGetIntDefaultValue(t *testing.T) {
	name := "int.variable"
	defaultValue := int64(12345)
	configManager = nil
	Config().Reset()

	testGetIntDefault(t, Config(), name, defaultValue)
}

func TestSetGetBool(t *testing.T) {
	name := "bool.variable"
	value := true
	configManager = nil
	Config().Reset()

	testSetGetBool(t, Config(), name, value)
}

func TestGetBoolDefaultValue(t *testing.T) {
	name := "bool.variable"
	defaultValue := true
	configManager = nil
	Config().Reset()
	testGetBoolDefault(t, Config(), name, defaultValue)
}
