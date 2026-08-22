package broker

import (
	"fmt"
	"pubsub/core"
	"pubsub/models"
	"sync"
)

// DefaultTopic implements core.Topic
// It delegates partitioning to the injected core.Partitioner.
type DefaultTopic struct {
	Name        string
	Partitions  []core.Partition
	partitioner core.Partitioner
	mu          sync.RWMutex
}

func NewTopic(name string, numPartitions int, p core.Partitioner) *DefaultTopic {
	partitions := make([]core.Partition, numPartitions)
	for i := 0; i < numPartitions; i++ {
		partitions[i] = NewMemoryPartition(fmt.Sprintf("%s-part-%d", name, i))
	}
	
	if p == nil {
		p = &DefaultPartitioner{}
	}
	
	return &DefaultTopic{
		Name:        name,
		Partitions:  partitions,
		partitioner: p,
	}
}

// Publish distributes the message to a partition using the Strategy pattern (Partitioner).
func (t *DefaultTopic) Publish(msg *models.Message) (int, uint64, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	numPart := len(t.Partitions)
	if numPart == 0 {
		return 0, 0, fmt.Errorf("no partitions available in topic")
	}

	partitionIndex := t.partitioner.Partition(msg, numPart)
	part := t.Partitions[partitionIndex]
	offset := part.Append(msg)
	
	return partitionIndex, offset, nil
}

// GetPartition retrieves a specific partition by index.
func (t *DefaultTopic) GetPartition(index int) core.Partition {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if index < 0 || index >= len(t.Partitions) {
		return nil
	}
	return t.Partitions[index]
}

// PartitionsCount returns the number of partitions.
func (t *DefaultTopic) PartitionsCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.Partitions)
}
