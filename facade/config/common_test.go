package facade

import (
	"testing"
)

func testSetGetString(t *testing.T, cf ConfigManager, name string, value string) {
	str, err := cf.GetString(name)
	if err != nil {
		t.Fatal(err)
	}
	if str != "" {
		t.Fatalf("Config is not empty on start, got %s", str)
	}

	err = cf.SetString(name, value)
	if err != nil {
		t.Fatal(err)
	}

	str, err = cf.GetString(name)
	if err != nil {
		t.Fatal(err)
	}
	if str != value {
		t.Errorf("Expected %s=%s, but got %s", name, value, str)
	}
}

func testGetStringDefault(t *testing.T, cf ConfigManager, name string, defaultValue string) {
	str, err := cf.GetString(name)
	if err != nil {
		t.Fatal(err)
	}
	if str != "" {
		t.Fatalf("Config is not empty on start")
	}

	str, err = cf.GetString(name, defaultValue)
	if err != nil {
		t.Fatal(err)
	}
	if str != defaultValue {
		t.Errorf("Expected %s=%s, but got %s", name, defaultValue, str)
	}
}

func testSetGetInt(t *testing.T, cf ConfigManager, name string, value int) {
	i, err := cf.GetInt(name)
	if err != nil {
		t.Fatal(err)
	}
	if i != 0 {
		t.Fatalf("Config is not empty on start")
	}

	err = cf.SetInt(name, value)
	if err != nil {
		t.Fatal(err)
	}
	i, err = cf.GetInt(name)
	if err != nil {
		t.Fatal(err)
	}
	if i != value {
		t.Errorf("Expected %s=%d, but got %d", name, value, i)
	}
}

func testGetIntDefault(t *testing.T, cf ConfigManager, name string, defaultValue int) {
	i, err := cf.GetInt(name)
	if err != nil {
		t.Fatal(err)
	}
	if i != 0 {
		t.Fatalf("Config is not empty on start")
	}
	i, err = cf.GetInt(name, defaultValue)
	if err != nil {
		t.Fatal(err)
	}
	if i != defaultValue {
		t.Errorf("Expected %s=%d, but got %d", name, defaultValue, i)
	}
}

func testSetGetBool(t *testing.T, cf ConfigManager, name string, value bool) {
	b, err := cf.GetBool(name)
	if err != nil {
		t.Fatal(err)
	}
	if b {
		t.Fatalf("Config is not empty on start")
	}

	err = cf.SetBool(name, value)
	if err != nil {
		t.Fatal(err)
	}

	b, err = cf.GetBool(name)
	if err != nil {
		t.Fatal(err)
	}
	if b != value {
		t.Errorf("Expected %s=%t, but got %t", name, value, b)
	}
}

func testGetBoolDefault(t *testing.T, cf ConfigManager, name string, defaultValue bool) {
	b, err := cf.GetBool(name)
	if err != nil {
		t.Fatal(err)
	}
	if b {
		t.Fatalf("Config is not empty on start")
	}

	b, err = cf.GetBool(name, defaultValue)
	if err != nil {
		t.Fatal(err)
	}
	if b != defaultValue {
		t.Errorf("Expected %s=%t, but got %t", name, defaultValue, b)
	}
}

func helperStringInSlice(s string, sl []string) bool {
	for _, val := range sl {
		if s == val {
			return true
		}
	}
	return false
}
