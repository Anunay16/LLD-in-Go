package broker

import (
	"pubsub/models"
	"sync"
)

// MemoryPartition implements core.Partition using in-memory slices.
type MemoryPartition struct {
	ID       string
	messages []models.Message
	mu       sync.RWMutex
	offset   uint64 // current max offset
}

func NewMemoryPartition(id string) *MemoryPartition {
	return &MemoryPartition{
		ID:       id,
		messages: make([]models.Message, 0),
		offset:   0,
	}
}

// Append adds a new message to the partition and returns its offset.
func (p *MemoryPartition) Append(msg *models.Message) uint64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	msg.Offset = p.offset
	p.messages = append(p.messages, *msg)
	currentOffset := p.offset
	p.offset++

	return currentOffset
}

// Read retrieves a message at a specific offset.
// Returns false if the offset is out of bounds (message not yet produced).
func (p *MemoryPartition) Read(offset uint64) (*models.Message, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if offset >= uint64(len(p.messages)) {
		return nil, false
	}
	
	msg := p.messages[offset]
	return &msg, true
}

// GetMaxOffset returns the current maximum offset (next expected message).
func (p *MemoryPartition) GetMaxOffset() uint64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.offset
}
