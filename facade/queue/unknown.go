package queue

import (
	"fmt"

	"http-attenuator/data"
)

type errorQueueManager struct {
	typeName QueueManagerType
	err      error
}

func NewUnknownQueue(typeName QueueManagerType, err error) QueueManager {
	return &errorQueueManager{
		typeName: typeName,
		err:      err,
	}
}

func (q *errorQueueManager) Type() QueueManagerType {
	return q.typeName
}

func (q *errorQueueManager) PutTopic(topic string, m data.Message, block bool) error {
	return fmt.Errorf("ERROR|facade/queue|queue error|%s", q.err)
}

func (q *errorQueueManager) FetchTopic(topic string, block bool) (data.Message, error) {
	return nil, fmt.Errorf("ERROR|facade/queue|queue error|%s", q.err)
}

func (q *errorQueueManager) CreateTopic(topic string) error {
	return fmt.Errorf("ERROR|facade/queue|queue error|%s", q.err)
}
