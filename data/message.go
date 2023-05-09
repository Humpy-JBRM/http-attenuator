package data

import (
	"fmt"

	"github.com/google/uuid"
)

// Message is the base data structure used for queueing and
// data pipelines etc
//
// The actual implementation is TBD
type Message interface {
	GetId() string
}

type BaseMessage struct {
	Id string `json:"id"`
}

func (m *BaseMessage) GetId() string {
	return m.Id
}

type MessageBuilder interface {
	MessageBuilderFromJson([]byte) (MessageBuilder, error)
	Id(id string) MessageBuilder
	Build() (Message, error)
}

type MessageBuilderImpl struct {
	impl BaseMessage
}

func NewMessageBuilder() MessageBuilder {
	return &MessageBuilderImpl{
		impl: BaseMessage{
			Id: uuid.NewString(),
		},
	}
}

func (b *MessageBuilderImpl) MessageBuilderFromJson([]byte) (MessageBuilder, error) {
	return b, fmt.Errorf("TODO(john): implement MessageBuilder.MessageBuilderFromJson()")
}

func (b *MessageBuilderImpl) Id(id string) MessageBuilder {
	b.impl.Id = id
	return b
}

func (b *MessageBuilderImpl) Build() (Message, error) {
	messageCopy := b.impl
	return &messageCopy, nil
}
