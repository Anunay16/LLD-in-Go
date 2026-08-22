package core

import "pubsub/models"

// Broker represents the central pub-sub system registry.
// Clients depend on this interface rather than the concrete implementation.
type Broker interface {
	CreateTopic(name string, numPartitions int) error
	GetTopic(name string) (Topic, error)
}

// Topic represents a logical stream of messages.
type Topic interface {
	Publish(msg *models.Message) (int, uint64, error)
	GetPartition(index int) Partition
	PartitionsCount() int
}

// Partition represents an ordered, immutable sequence of messages.
// This interface allows abstracting away the storage mechanism (Memory vs Disk).
type Partition interface {
	Append(msg *models.Message) uint64
	Read(offset uint64) (*models.Message, bool)
	GetMaxOffset() uint64
}

// Partitioner defines the strategy for assigning messages to partitions.
// This follows the Open/Closed Principle (OCP) and Strategy Pattern, 
// allowing new partition logic without modifying the Topic.
type Partitioner interface {
	Partition(msg *models.Message, numPartitions int) int
}
