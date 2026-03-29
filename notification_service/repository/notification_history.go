package repository

import (
	"notification-service/types"
	"sync"
)

type NotificationHistoryRepository interface {
	Save(history types.NotificationHistory)
	GetAll() []types.NotificationHistory
}

type InMemoryHistoryRepository struct {
	mu      sync.Mutex
	records []types.NotificationHistory
}

func NewInMemoryHistoryRepository() *InMemoryHistoryRepository {
	return &InMemoryHistoryRepository{
		records: []types.NotificationHistory{},
	}
}

func (r *InMemoryHistoryRepository) Save(history types.NotificationHistory) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.records = append(r.records, history)
}

func (r *InMemoryHistoryRepository) GetAll() []types.NotificationHistory {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.records
}
