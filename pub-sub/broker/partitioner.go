package broker

import (
	"hash/fnv"
	"pubsub/models"
)

// DefaultPartitioner implements core.Partitioner
// It uses hash-based partitioning if a key is present, otherwise timestamp-based.
type DefaultPartitioner struct{}

func (dp *DefaultPartitioner) Partition(msg *models.Message, numPartitions int) int {
	if numPartitions == 0 {
		return 0
	}
	
	if len(msg.Key) > 0 {
		h := fnv.New32a()
		h.Write(msg.Key)
		return int(h.Sum32() % uint32(numPartitions))
	}
	
	// Fallback for keyless messages
	return int(msg.Timestamp.UnixNano() % int64(numPartitions))
}
