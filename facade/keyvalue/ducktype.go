package facade

type KeyValue interface {
	Set(key string, value interface{}) error
	GetString(key string) (string, error)
	GetInt(key string) (int64, error)
	GetFloat(key string) (float64, error)
	GetBool(key string) (bool, error)
	Delete(key string) error
	Add(key string, delta int64) (int64, error)
	Dec(key string, delta int64) (int64, error)
}
