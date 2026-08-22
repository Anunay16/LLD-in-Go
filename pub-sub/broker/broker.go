package broker

import (
	"errors"
	"pubsub/core"
	"sync"
)

// MemoryBroker implements core.Broker
type MemoryBroker struct {
	topics map[string]core.Topic
	mu     sync.RWMutex
}

var (
	instance core.Broker
	once     sync.Once
)

// GetBroker implements the Singleton pattern for the broker, returning the core.Broker interface.
func GetBroker() core.Broker {
	once.Do(func() {
		instance = &MemoryBroker{
			topics: make(map[string]core.Topic),
		}
	})
	return instance
}

// CreateTopic initializes a new topic with the specified number of partitions.
func (b *MemoryBroker) CreateTopic(name string, numPartitions int) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.topics[name]; exists {
		return errors.New("topic already exists")
	}

	b.topics[name] = NewTopic(name, numPartitions, nil)
	return nil
}

// GetTopic retrieves an existing topic by name.
func (b *MemoryBroker) GetTopic(name string) (core.Topic, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	topic, exists := b.topics[name]
	if !exists {
		return nil, errors.New("topic not found")
	}
	return topic, nil
}
