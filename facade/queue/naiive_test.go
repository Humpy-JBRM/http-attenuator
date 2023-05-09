package queue

import (
	"testing"

	"http-attenuator/data"
)

func TestPutAndFetchTopic(t *testing.T) {
	// Create a message
	topic := "theTopic"
	msg, err := data.NewMessageBuilder().Build()
	if err != nil {
		t.Fatal(err)
	}

	// Put the message on the queue
	err = Queue().PutTopic(topic, msg, false)
	if err != nil {
		t.Fatal(err)
	}

	// Get the message from the topic queue
	fetched, err := Queue().FetchTopic(topic, false)
	if err != nil {
		t.Fatal(err)
	}
	if fetched == nil {
		t.Errorf("Expected to fetch a message from queue '%s'", topic)
	}

	if fetched.GetId() != msg.GetId() {
		t.Errorf("Expected to fetch message id %s, but got %s", msg.GetId(), fetched.GetId())
	}
}
