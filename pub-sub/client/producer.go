package client

import (
	"fmt"
	"pubsub/core"
	"pubsub/models"
	"time"
)

// Producer handles publishing messages to topics in the broker.
// It depends on the core.Broker interface, adhering to DIP.
type Producer struct {
	broker core.Broker
}

// NewProducer creates a new producer instance.
func NewProducer(b core.Broker) *Producer {
	return &Producer{broker: b}
}

// Publish sends a message to the specified topic.
// It returns the partition index it was written to and the offset.
func (p *Producer) Publish(topicName string, key []byte, value []byte) (int, uint64, error) {
	topic, err := p.broker.GetTopic(topicName)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get topic: %w", err)
	}

	msg := &models.Message{
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
	}

	return topic.Publish(msg)
}
