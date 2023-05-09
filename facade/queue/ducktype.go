package queue

import (
	"http-attenuator/data"
)

// QueueManager encapsulates the behaviour that we want from a
// FIFO / priority queue manager component, abstracting away the
// implementation.
//
// The naiive implementation runs entirely in memory, and has no
// external dependencies.
//
// A production implementation would be a facade for Artemis / Kafka / Redis
type QueueManager interface {
	Type() QueueManagerType
	PutTopic(topic string, m data.Message, block bool) error
	FetchTopic(topic string, block bool) (data.Message, error)
	CreateTopic(topic string) error
}

// QueueManagerFactory produces QueueManager instances based on
// the criteria we set.
//
// This is how we can flip implementations without refactoring
// any code
type QueueManagerFactory interface {
	SetType(qmType QueueManagerType) QueueManagerFactory
	SetAttenuator(attenuator data.Attenuator) QueueManagerFactory
	SetRoot(root string) QueueManagerFactory
	New() (QueueManager, error)
}
