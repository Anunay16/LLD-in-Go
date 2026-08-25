package store

import (
	"context"
	"sort"
	"sync"
	"time"

	"job_scheduler/pkg/model"
)

// MemoryJobStore is an in-memory thread-safe implementation of JobStore.
type MemoryJobStore struct {
	mu   sync.RWMutex
	jobs map[string]*model.Job
}

func NewMemoryJobStore() *MemoryJobStore {
	return &MemoryJobStore{
		jobs: make(map[string]*model.Job),
	}
}

func (s *MemoryJobStore) Create(ctx context.Context, job *model.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[job.ID]; exists {
		return ErrJobExists
	}

	// Store a copy / pointer
	s.jobs[job.ID] = job
	return nil
}

func (s *MemoryJobStore) Get(ctx context.Context, id string) (*model.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[id]
	if !exists {
		return nil, ErrJobNotFound
	}
	return job, nil
}

func (s *MemoryJobStore) Update(ctx context.Context, job *model.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[job.ID]; !exists {
		return ErrJobNotFound
	}

	job.UpdatedAt = time.Now()
	s.jobs[job.ID] = job
	return nil
}

func (s *MemoryJobStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[id]; !exists {
		return ErrJobNotFound
	}

	delete(s.jobs, id)
	return nil
}

func (s *MemoryJobStore) GetDueJobs(ctx context.Context, until time.Time, limit int) ([]*model.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var dueJobs []*model.Job
	for _, job := range s.jobs {
		if job.State == model.JobStatePending && job.NextRunAt != nil && !job.NextRunAt.After(until) {
			dueJobs = append(dueJobs, job)
		}
	}

	// Sort by NextRunAt ascending for priority
	sort.Slice(dueJobs, func(i, j int) bool {
		return dueJobs[i].NextRunAt.Before(*dueJobs[j].NextRunAt)
	})

	if limit > 0 && len(dueJobs) > limit {
		dueJobs = dueJobs[:limit]
	}

	return dueJobs, nil
}
