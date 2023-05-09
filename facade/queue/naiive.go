package queue

import (
	"fmt"
	"sync"
	"time"

	"http-attenuator/data"
)

type naiiveQueueManagerImpl struct {
	topics     map[string]chan data.Message
	mutex      sync.Mutex
	attenuator data.Attenuator
}

func (q *naiiveQueueManagerImpl) Type() QueueManagerType {
	return QUEUE_NAIIVE
}

func (q *naiiveQueueManagerImpl) PutTopic(topic string, m data.Message, block bool) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if _, exists := q.topics[NormaliseQueueName(topic)]; !exists {
		q.topics[NormaliseQueueName(topic)] = make(chan data.Message, 1000)
	}

	q.topics[NormaliseQueueName(topic)] <- m
	return nil
}

func (q *naiiveQueueManagerImpl) FetchTopic(topic string, block bool) (data.Message, error) {
	if _, exists := q.topics[NormaliseQueueName(topic)]; !exists {
		return nil, fmt.Errorf("ERROR|facade/queue/naiive|Topic %s does not exist|", topic)
	}

	if q.attenuator != nil {
		now := time.Now()
		q.attenuator.Wait()
		waitTime := (time.Now().UnixNano() - now.UnixNano()/1000000)
		queueAttenuatedWaitMillis.WithLabelValues(string(QUEUE_NAIIVE)).Add(float64(waitTime))
	}

	if block {
		return <-q.topics[NormaliseQueueName(topic)], nil
	}

	select {
	case m := <-q.topics[NormaliseQueueName(topic)]:
		return m, nil

	default:
		return nil, nil
	}
}

func (q *naiiveQueueManagerImpl) CreateTopic(topic string) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if _, exists := q.topics[NormaliseQueueName(topic)]; !exists {
		q.topics[NormaliseQueueName(topic)] = make(chan data.Message, 1000)
	}
	return nil
}
