package model

import (
	"time"

	"job_scheduler/pkg/strategy"
)

// JobState represents the lifecycle state of a Job.
type JobState string

const (
	JobStatePending   JobState = "PENDING"
	JobStateRunning   JobState = "RUNNING"
	JobStateCompleted JobState = "COMPLETED"
	JobStateFailed    JobState = "FAILED"
)

// Job represents a schedulable unit of work.
type Job struct {
	ID           string
	Name         string
	State        JobState
	NextRunAt    *time.Time
	Schedule     strategy.ScheduleStrategy
	Task         strategy.TaskStrategy
	RetryPolicy  strategy.RetryPolicyStrategy
	CurrentRetry int
	LastRunAt    *time.Time
	LastError    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewJob creates a new Job instance with initial state and timestamps.
func NewJob(
	id, name string,
	schedule strategy.ScheduleStrategy,
	task strategy.TaskStrategy,
	retryPolicy strategy.RetryPolicyStrategy,
) *Job {
	now := time.Now()
	var nextRun *time.Time
	if schedule != nil {
		nextRun = schedule.NextRun(now)
	}

	return &Job{
		ID:           id,
		Name:         name,
		State:        JobStatePending,
		NextRunAt:    nextRun,
		Schedule:     schedule,
		Task:         task,
		RetryPolicy:  retryPolicy,
		CurrentRetry: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
