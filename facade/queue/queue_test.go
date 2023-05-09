package queue

import (
	"testing"
)

func TestQueueSingleton(t *testing.T) {
	if Queue() != Queue() {
		t.Errorf("Queue() singleton does not return the same pointer")
	}
}

func TestQueueFactorySingleton(t *testing.T) {
	if QueueFactory() != QueueFactory() {
		t.Errorf("QueueFactory() singleton does not return the same pointer")
	}
}

func TestQueueFactoryReturnsNaiiveByDefault(t *testing.T) {
	q, err := QueueFactory().New()
	if err != nil {
		t.Error(err)
	}
	if q == nil {
		t.Errorf("Expected a QueueManager, but got nil")
	}
	if q.Type() != QUEUE_NAIIVE {
		t.Errorf("Expected QueueManager type '%s', but got '%s'", QUEUE_NAIIVE, q.Type())
	}
}

func TestQueueFactoryPukesOnUnknownType(t *testing.T) {
	_, err := QueueFactory().SetType("unknownType").New()
	if err == nil {
		t.Errorf("Expected an 'unknown type' error")
	}
}

func TestQueueInstanceIsUnknownWhenUnknownType(t *testing.T) {
	queueManager = nil
	QueueFactory().SetType("unknownType")

	_, err := Queue().FetchTopic("", false)
	if err == nil {
		t.Errorf("Expected an error because the queue implementation is unknown")
	}
}
