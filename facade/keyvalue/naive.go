package facade

import (
	"fmt"
	"reflect"
	"sync"
)

type NaiveKeyValue struct {
	valuesByName map[string]any
	mutex        sync.RWMutex
}

func NewNaiveKeyValue() (KeyValue, error) {
	return &NaiveKeyValue{
		valuesByName: make(map[string]any),
	}, nil
}

func (kv *NaiveKeyValue) Set(key string, value any) error {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()
	kv.valuesByName[key] = value
	return nil
}

func (kv *NaiveKeyValue) GetString(key string) (string, error) {
	kv.mutex.RLock()
	defer kv.mutex.RUnlock()
	v := kv.valuesByName[key]
	if v == nil {
		return "", nil
	}
	return fmt.Sprint(v), nil
}

func (kv *NaiveKeyValue) GetInt(key string) (int64, error) {
	kv.mutex.RLock()
	defer kv.mutex.RUnlock()
	v := kv.valuesByName[key]
	if v == nil {
		return 0, nil
	}

	switch v.(type) {
	case int:
		return int64(v.(int)), nil
	case int64:
		return v.(int64), nil
	case float32:
		return int64(v.(float32)), nil
	case float64:
		return int64(v.(float64)), nil
	default:
		return 0, fmt.Errorf("GetInt(%s): cannot convert %s to int64", key, reflect.TypeOf(v))
	}
}

func (kv *NaiveKeyValue) GetFloat(key string) (float64, error) {
	kv.mutex.RLock()
	defer kv.mutex.RUnlock()
	v := kv.valuesByName[key]
	if v == nil {
		return 0, nil
	}

	switch v.(type) {
	case int:
		return float64(v.(int)), nil
	case int64:
		return float64(v.(int64)), nil
	case float32:
		return float64(v.(float32)), nil
	case float64:
		return v.(float64), nil
	default:
		return 0, fmt.Errorf("GetInt(%s): cannot convert %s to int64", key, reflect.TypeOf(v))
	}
}

func (kv *NaiveKeyValue) GetBool(key string) (bool, error) {
	kv.mutex.RLock()
	defer kv.mutex.RUnlock()
	v := kv.valuesByName[key]
	if v == nil {
		return false, nil
	}

	return fmt.Sprint(v) == "true", nil
}

func (kv *NaiveKeyValue) Delete(key string) error {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()
	delete(kv.valuesByName, key)
	return nil
}

func (kv *NaiveKeyValue) Add(key string, delta int64) (int64, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()
	return 0, nil
}

func (kv *NaiveKeyValue) Dec(key string, delta int64) (int64, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()
	return 0, nil
}
