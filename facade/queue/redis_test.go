package queue

import (
	"fmt"
	"testing"

	config "http-attenuator/facade/config"

	"http-attenuator/data"
)

func TestRedisQueue(t *testing.T) {
	// config.Config().SetString(util.CONF_QUEUE_IMPL, "redis")
	// config.Config().SetString(util.CONF_REDIS_URL, "127.0.0.1:6379")
	// redis.InitialiseRedis()
	if url, _ := config.Config().GetString(data.CONF_REDIS_HOST); url == "" {
		t.Skip()
	}

	// Get a redis queue
	q, err := QueueFactory().New()
	if err != nil {
		t.Fatal(err)
	}
	if q == nil {
		t.Errorf("Expected a QueueManager, but got nil")
	}
	if q.Type() != QUEUE_REDIS {
		t.Errorf("Expected QueueManager type '%s', but got '%s'", QUEUE_REDIS, q.Type())
	}

	_, err = data.NewMessageBuilder().Build()
	if err != nil {
		t.Fatal(err)
	}

	// Clean the queue
	topic := "topic"
	for item, _ := q.FetchTopic(topic, false); item != nil; {
		item, _ = q.FetchTopic(topic, false)
	}

	// Put a few messages on the queue
	expectedMessages := 5
	for i := 1; i <= expectedMessages; i++ {
		mb := data.NewMessageBuilder().Id(fmt.Sprintf("msg-id-%d", i))
		m, err := mb.Build()
		if err != nil {
			t.Fatal(err)
		}
		err = q.PutTopic(topic, m, false)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Make sure we get the correct number of messages
	messages := make([]data.Message, 0)
	var nextMessage data.Message
	for nextMessage, err = q.FetchTopic(topic, false); nextMessage != nil; {
		if err != nil {
			t.Fatal(err)
		}
		messages = append(messages, nextMessage)
		nextMessage, err = q.FetchTopic(topic, false)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != expectedMessages {
		t.Errorf("Expected %d messages, got %d", expectedMessages, len(messages))
	}
	for i, msg := range messages {
		expectedId := fmt.Sprintf("msg-id-%d", i)
		if msg.GetId() != expectedId {
			t.Errorf("Expected msgId=%s, got %s", expectedId, msg.GetId())
		}
	}
}
