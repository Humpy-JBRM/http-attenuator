package queue

import (
	"encoding/json"
	"fmt"
	"http-attenuator/data"
	redis "http-attenuator/facade/redis"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var queueOperationsStarted = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "atat",
	Name:      "queue_op_started",
	Help:      "The number of operations performed",
},
	[]string{"function"},
)
var queueOperationsSuccess = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "atat",
	Name:      "queue_op_success",
	Help:      "The number of sucesses",
},
	[]string{"function"},
)
var queueOperationsError = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "atat",
	Name:      "queue_op_error",
	Help:      "The number of errors",
},
	[]string{"function"},
)
var messagesQueued = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "atat",
	Name:      "messages_queued",
	Help:      "The number of messages queued (keyed by topic)",
},
	[]string{"topic"},
)
var messagesfetched = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "atat",
	Name:      "messages_fetched",
	Help:      "The number of messages fetched from a queue (keyed by topic)",
},
	[]string{"topic"},
)

type redisQueueManager struct {
	attenuator data.Attenuator
}

const DEFAULT_TIMEOUT float64 = 0.1 // value in seconds

func (q *redisQueueManager) Type() QueueManagerType {
	return QUEUE_REDIS
}

func (q *redisQueueManager) PutTopic(topic string, msg data.Message, block bool) (err error) {
	msgJson, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("ERROR|redis.PutTopic(%s)|Could not add message to queue|%s", topic, err.Error())
	}
	_, err = redis.QueueDo(
		"LPUSH",
		topic,
		string(msgJson),
	)
	if err != nil {
		return fmt.Errorf("ERROR|redis.PutTopic(%s)|Could not add message to queue|%s", topic, err.Error())
	}

	return err
}

func (q *redisQueueManager) FetchTopic(topic string, block bool) (data.Message, error) {
	queueOperationsStarted.WithLabelValues("FetchTopic").Inc()

	if q.attenuator != nil {
		now := time.Now()
		q.attenuator.Wait()
		waitTime := (time.Now().UnixNano() - now.UnixNano()/1000000)
		queueAttenuatedWaitMillis.WithLabelValues(string(QUEUE_REDIS)).Add(float64(waitTime))
	}
	timeout := DEFAULT_TIMEOUT

	var err error
	var entry interface{}

	if block {
		// BLOCKING
		timeout = float64(0)
		entry, err = redis.QueueDo(
			"BRPOP",
			topic,
			timeout,
		)
		// BRPOP returns a slice [topic, item]
		switch entry.(type) {
		case []interface{}:
			entry = entry.([]interface{})[1]
		}
	} else {
		// Non-blocking
		entry, err = redis.QueueDo(
			"RPOP",
			topic,
		)
	}
	if err != nil {
		queueOperationsError.WithLabelValues("FetchTopic").Inc()
		return nil, err
	}

	switch entry.(type) {
	case []byte:
		rawJson := entry.([]byte)
		builder, err := data.NewMessageBuilder().MessageBuilderFromJson(rawJson)
		if err != nil {
			queueOperationsError.WithLabelValues("FetchTopic").Inc()
			return nil, fmt.Errorf("ERROR|redis.FetchTopic(%s)|Could not unmarshall message from queue|%s", topic, err.Error())
		}
		msg, err := builder.Build()
		if err != nil {
			queueOperationsError.WithLabelValues("FetchTopic").Inc()
			return nil, fmt.Errorf("ERROR|redis.FetchTopic(%s)|Could not unmarshall message from queue|%s", topic, err.Error())
		}
		queueOperationsSuccess.WithLabelValues("FetchTopic").Inc()
		return msg, nil
	}
	queueOperationsSuccess.WithLabelValues("FetchTopic").Inc()
	return nil, nil
}

func (q *redisQueueManager) CreateTopic(topic string) error {
	// This is a NOP in redis
	return nil
}
