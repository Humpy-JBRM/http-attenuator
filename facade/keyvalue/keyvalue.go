package facade

import (
	"fmt"
	"http-attenuator/data"
	config "http-attenuator/facade/config"
	"sync"
)

// the kv store is a singleton.  When you consider that it can also
// have a redis implementation - so that state can be shared across
// multiple instances / pods - then this makes intuitive sense.
var instance KeyValue
var keyValueOnce sync.Once

func GetKeyValue() (KeyValue, error) {
	var err error
	keyValueOnce.Do(func() {
		impl, e := config.Config().GetString(data.CONF_KEYVALUE_IMPL)
		if e != nil {
			err = e
			return
		}

		switch impl {
		case "naive", "":
			instance, err = NewNaiveKeyValue()

		case "redis":
			instance, err = NewRedisKeyValue()

		default:
			err = fmt.Errorf("KeyValue implementation '%s' not implemented", impl)
		}
	})

	return instance, err
}
