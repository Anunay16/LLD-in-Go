package store

import (
	"context"
	"errors"
	"time"

	"job_scheduler/pkg/model"
)

var (
	ErrJobNotFound = errors.New("job not found")
	ErrJobExists   = errors.New("job already exists")
)

// JobStore defines the repository interface for persisting and retrieving jobs.
type JobStore interface {
	Create(ctx context.Context, job *model.Job) error
	Get(ctx context.Context, id string) (*model.Job, error)
	Update(ctx context.Context, job *model.Job) error
	Delete(ctx context.Context, id string) error
	GetDueJobs(ctx context.Context, until time.Time, limit int) ([]*model.Job, error)
}
