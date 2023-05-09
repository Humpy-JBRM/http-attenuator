package queue

// A facade which encapsulates a SQL queue.
//
// The naiive implementation is an in-memory QUEUE
import (
	"fmt"
	redis "http-attenuator/facade/redis"
	"strings"
	"sync"

	"http-attenuator/data"
	config "http-attenuator/facade/config"
	"http-attenuator/util"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var queueAttenuatedWaitMillis = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "faultmonkey",
	Name:      "queue_attenuated_wait_millis",
	Help:      "The number of milliseconds we spend waiting for attenuation",
},
	[]string{"type"},
)

var queueManagerOnce sync.Once
var queueManager QueueManager

const DEFAULT_QUEUE string = ""

func initialise() {
	qm, err := QueueFactory().New()
	if err != nil {
		if strings.Index(err.Error(), "ERROR|") == 0 {
			util.DoError(err)
		} else {
			util.DoError(fmt.Errorf("ERROR|facade/queue.initialise()|Could not create instance|%s", err.Error()))
		}
		qm = &errorQueueManager{
			typeName: QUEUE_UNKNOWN,
		}
	}
	queueManager = qm
}

func Queue() QueueManager {
	queueManagerOnce.Do(initialise)
	if queueManager == nil {
		initialise()
	}
	return queueManager
}

type QueueManagerType string

const (
	QUEUE_NAIIVE  QueueManagerType = "naiive"
	QUEUE_REDIS   QueueManagerType = "redis"
	QUEUE_SMARTFS QueueManagerType = "smartfs"
	QUEUE_UNKNOWN QueueManagerType = "error"

	// Metadata names
	KV_METADATA_SERVER    string = "server"
	KV_METADATA_PORT      string = "port"
	KV_METADATA_MAXIDLE   string = "maxidle"
	KV_METADATA_MAXACTIVE string = "maxactive"
)

type queueManagerFactoryImpl struct {
	root       string
	qType      QueueManagerType
	attenuator data.Attenuator
}

var queueManagerFactoryOnce sync.Once
var queueManagerFactory QueueManagerFactory

func QueueFactory() QueueManagerFactory {
	queueManagerFactoryOnce.Do(func() {
		queueManagerFactory = &queueManagerFactoryImpl{}

		if qType, _ := config.Config().GetString(data.CONF_QUEUE_IMPL); qType == string(QUEUE_REDIS) {
			if url, _ := config.Config().GetString(data.CONF_REDIS_HOST); url != "" {
				queueManagerFactory.(*queueManagerFactoryImpl).qType = QUEUE_REDIS
				util.DoDebug(fmt.Sprintf("QueueFactory(): Setting type to %s (%s)", QUEUE_REDIS, url))
			}
		}
	})
	return queueManagerFactory
}

func (qf *queueManagerFactoryImpl) SetType(qType QueueManagerType) QueueManagerFactory {
	qf.qType = qType
	return qf
}

func (qf *queueManagerFactoryImpl) SetRoot(root string) QueueManagerFactory {
	qf.root = root
	return qf
}

func (qf *queueManagerFactoryImpl) SetAttenuator(attenuator data.Attenuator) QueueManagerFactory {
	qf.attenuator = attenuator
	return qf
}

func (qf *queueManagerFactoryImpl) New() (QueueManager, error) {
	switch qf.qType {
	case QUEUE_NAIIVE, "":
		return &naiiveQueueManagerImpl{
			topics: make(map[string]chan data.Message),
		}, nil

	case QUEUE_REDIS:
		redis.InitialiseRedisQueue()
		return &redisQueueManager{}, nil

	default:
		err := fmt.Errorf("ERROR|facade/queue|No implementation for queue manager type '%s'", qf.qType)
		return NewUnknownQueue(QUEUE_UNKNOWN, err), err
	}
}
