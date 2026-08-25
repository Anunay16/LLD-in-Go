package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"job_scheduler/pkg/executor"
	"job_scheduler/pkg/lock"
	"job_scheduler/pkg/model"
	"job_scheduler/pkg/store"
)

// Scheduler coordinates JobStore, Lock, and Executor to run scheduled jobs.
type Scheduler struct {
	store        store.JobStore
	lock         lock.Locker
	executor     executor.Executor
	pollInterval time.Duration
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// NewScheduler creates a Scheduler instance with its 3 core dependencies.
func NewScheduler(
	store store.JobStore,
	lock lock.Locker,
	executor executor.Executor,
	pollInterval time.Duration,
) *Scheduler {
	return &Scheduler{
		store:        store,
		lock:         lock,
		executor:     executor,
		pollInterval: pollInterval,
		stopChan:     make(chan struct{}),
	}
}

// AddJob registers a new job into the scheduler's store.
func (s *Scheduler) AddJob(ctx context.Context, job *model.Job) error {
	return s.store.Create(ctx, job)
}

// Start begins the polling loop in a background goroutine.
func (s *Scheduler) Start(ctx context.Context) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-s.stopChan:
				return
			case <-ticker.C:
				s.pollAndExecute(ctx)
			}
		}
	}()
}

// Stop gracefully terminates the scheduler polling loop.
func (s *Scheduler) Stop() {
	close(s.stopChan)
	s.wg.Wait()
}

func (s *Scheduler) pollAndExecute(ctx context.Context) {
	dueJobs, err := s.store.GetDueJobs(ctx, time.Now(), 10)
	if err != nil {
		return
	}

	for _, job := range dueJobs {
		go s.executeJob(ctx, job)
	}
}

func (s *Scheduler) executeJob(ctx context.Context, job *model.Job) {
	lockKey := fmt.Sprintf("job_lock:%s", job.ID)
	acquired, err := s.lock.Acquire(ctx, lockKey, 5*time.Second)
	if err != nil || !acquired {
		return
	}
	defer s.lock.Release(ctx, lockKey)

	// Fetch freshest state
	latestJob, err := s.store.Get(ctx, job.ID)
	if err != nil || latestJob.State == model.JobStateRunning {
		return
	}

	latestJob.State = model.JobStateRunning
	_ = s.store.Update(ctx, latestJob)

	now := time.Now()
	latestJob.LastRunAt = &now

	// Execute via Executor
	execErr := s.executor.Execute(ctx, latestJob)

	if execErr == nil {
		latestJob.LastError = ""
		latestJob.CurrentRetry = 0
		nextRun := latestJob.Schedule.NextRun(now)
		if nextRun != nil {
			latestJob.State = model.JobStatePending
			latestJob.NextRunAt = nextRun
		} else {
			latestJob.State = model.JobStateCompleted
			latestJob.NextRunAt = nil
		}
	} else {
		latestJob.LastError = execErr.Error()
		delay, shouldRetry := latestJob.RetryPolicy.NextRetry(latestJob.CurrentRetry + 1)
		if shouldRetry {
			latestJob.CurrentRetry++
			retryTime := now.Add(delay)
			latestJob.State = model.JobStatePending
			latestJob.NextRunAt = &retryTime
			fmt.Printf("[Scheduler] Job %s failed, retrying in %v (attempt %d)...\n",
				latestJob.Name, delay, latestJob.CurrentRetry)
		} else {
			nextRun := latestJob.Schedule.NextRun(now)
			if nextRun != nil {
				latestJob.State = model.JobStatePending
				latestJob.NextRunAt = nextRun
				latestJob.CurrentRetry = 0
			} else {
				latestJob.State = model.JobStateFailed
				latestJob.NextRunAt = nil
			}
			fmt.Printf("[Scheduler] Job %s permanently failed: %v\n", latestJob.Name, execErr)
		}
	}

	_ = s.store.Update(ctx, latestJob)
}
