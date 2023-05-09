package facade

import (
	"os"
	"testing"

	config "http-attenuator/facade/config"
)

func TestRedisKVCrud(t *testing.T) {
	t.Skip()
	os.Setenv("CONFIG_FILE", "../../config.yml")
	config.Config()

	kv, err := NewRedisKeyValue()
	if err != nil {
		t.Fatal(err)
	}

	key := "redis_kv_test"
	err = kv.Delete(key)
	if err != nil {
		t.Fatal(err)
	}

	expectedValue := "0"
	err = kv.Set(key, expectedValue)
	if err != nil {
		t.Error(err)
	}
	actualValue, err := kv.GetString(key)
	if err != nil {
		t.Error(err)
	}
	if expectedValue != actualValue {
		t.Errorf("Expected %s='%s', got '%s'", key, expectedValue, actualValue)
	}

	// Increment
	kv.Add(key, 10)
	expectedValue = "10"
	actualValue, err = kv.GetString(key)
	if err != nil {
		t.Error(err)
	}
	if expectedValue != actualValue {
		t.Errorf("Expected %s='%s', got '%s'", key, expectedValue, actualValue)
	}

	// Decrement
	kv.Dec(key, 5)
	expectedValue = "5"
	actualValue, err = kv.GetString(key)
	if err != nil {
		t.Error(err)
	}
	if expectedValue != actualValue {
		t.Errorf("Expected %s='%s', got '%s'", key, expectedValue, actualValue)
	}

	// GetInt
	actualInt, err := kv.GetInt(key)
	expectedInt := int64(5)
	if err != nil {
		t.Error(err)
	}
	if expectedInt != actualInt {
		t.Errorf("Expected %s='%d', got '%d'", key, expectedInt, actualInt)
	}

	// GetFloat
	actualFloat, err := kv.GetFloat(key)
	expectedFloat := float64(5)
	if err != nil {
		t.Error(err)
	}
	if expectedFloat != actualFloat {
		t.Errorf("Expected %s='%f', got '%f'", key, expectedFloat, actualFloat)
	}
}
