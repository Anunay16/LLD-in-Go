package client

import (
	"context"
	"fmt"
	"pubsub/core"
	"sync"
	"time"
)

// ConsumerGroup coordinates message consumption among multiple workers (consumers).
// It depends on the core.Broker interface, adhering to DIP.
type ConsumerGroup struct {
	GroupID string
	broker  core.Broker
	offsets map[string]map[int]uint64 // topic -> partition -> committed offset
	mu      sync.RWMutex
}

// NewConsumerGroup creates a new ConsumerGroup instance.
func NewConsumerGroup(id string, b core.Broker) *ConsumerGroup {
	return &ConsumerGroup{
		GroupID: id,
		broker:  b,
		offsets: make(map[string]map[int]uint64),
	}
}

// Subscribe starts the consumer group workers for a given topic.
func (cg *ConsumerGroup) Subscribe(ctx context.Context, topicName string, handler func(partition int, key []byte, value []byte)) error {
	topic, err := cg.broker.GetTopic(topicName)
	if err != nil {
		return err
	}

	// Initialize offsets for topic if not exists
	cg.mu.Lock()
	if _, ok := cg.offsets[topicName]; !ok {
		cg.offsets[topicName] = make(map[int]uint64)
	}
	cg.mu.Unlock()

	numPartitions := topic.PartitionsCount()
	if numPartitions == 0 {
		return fmt.Errorf("topic %s has no partitions", topicName)
	}

	for i := 0; i < numPartitions; i++ {
		go cg.consumePartition(ctx, topicName, i, topic.GetPartition(i), handler)
	}

	return nil
}

func (cg *ConsumerGroup) consumePartition(ctx context.Context, topicName string, partIdx int, partition core.Partition, handler func(partition int, key []byte, value []byte)) {
	cg.mu.RLock()
	currentOffset := cg.offsets[topicName][partIdx]
	cg.mu.RUnlock()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, found := partition.Read(currentOffset)
			if found {
				handler(partIdx, msg.Key, msg.Value)
				currentOffset++
				
				// Commit offset (simulated persistent storage)
				cg.mu.Lock()
				cg.offsets[topicName][partIdx] = currentOffset
				cg.mu.Unlock()
			} else {
				// No new messages, yield/sleep before polling again
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}
