package booking

import (
	"context"
	"fmt"
	"sync"
)

// Repository handles persistence operations for Shows and Holds.
type Repository interface {
	GetShow(ctx context.Context, id string) (*Show, error)
	SaveShow(ctx context.Context, show *Show) error
	CreateHold(ctx context.Context, hold *Hold) error
	GetHold(ctx context.Context, id string) (*Hold, error)
	UpdateHold(ctx context.Context, hold *Hold) error
}

// InMemoryRepository is a thread-safe in-memory repository implementation.
type InMemoryRepository struct {
	mu    sync.RWMutex
	shows map[string]*Show
	holds map[string]*Hold
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		shows: make(map[string]*Show),
		holds: make(map[string]*Hold),
	}
}

func (r *InMemoryRepository) GetShow(ctx context.Context, id string) (*Show, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	show, exists := r.shows[id]
	if !exists {
		return nil, ErrShowNotFound
	}

	// Deep copy to prevent caller mutation outside locking
	return cloneShow(show), nil
}

func (r *InMemoryRepository) SaveShow(ctx context.Context, show *Show) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.shows[show.ID] = cloneShow(show)
	return nil
}

func (r *InMemoryRepository) CreateHold(ctx context.Context, hold *Hold) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.holds[hold.ID]; exists {
		return fmt.Errorf("hold with id %s already exists", hold.ID)
	}

	r.holds[hold.ID] = cloneHold(hold)
	return nil
}

func (r *InMemoryRepository) GetHold(ctx context.Context, id string) (*Hold, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	hold, exists := r.holds[id]
	if !exists {
		return nil, ErrHoldNotFound
	}
	return cloneHold(hold), nil
}

func (r *InMemoryRepository) UpdateHold(ctx context.Context, hold *Hold) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.holds[hold.ID]; !exists {
		return ErrHoldNotFound
	}

	r.holds[hold.ID] = cloneHold(hold)
	return nil
}

// Helper deep copy functions to prevent data races on returned pointers
func cloneShow(s *Show) *Show {
	seatsCopy := make(map[string]*Seat, len(s.Seats))
	for k, v := range s.Seats {
		seatCopy := *v
		seatsCopy[k] = &seatCopy
	}
	return &Show{
		ID:    s.ID,
		Title: s.Title,
		Seats: seatsCopy,
	}
}

func cloneHold(h *Hold) *Hold {
	seatIDsCopy := make([]string, len(h.SeatIDs))
	copy(seatIDsCopy, h.SeatIDs)
	return &Hold{
		ID:        h.ID,
		ShowID:    h.ShowID,
		UserID:    h.UserID,
		SeatIDs:   seatIDsCopy,
		ExpiresAt: h.ExpiresAt,
		Status:    h.Status,
	}
}
