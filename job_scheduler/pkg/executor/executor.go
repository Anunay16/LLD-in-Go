package executor

import (
	"context"

	"job_scheduler/pkg/model"
)

// Executor defines the contract for executing a job's task.
type Executor interface {
	Execute(ctx context.Context, job *model.Job) error
}

// DefaultExecutor runs the job's task strategy directly.
type DefaultExecutor struct{}

func NewDefaultExecutor() *DefaultExecutor {
	return &DefaultExecutor{}
}

func (e *DefaultExecutor) Execute(ctx context.Context, job *model.Job) error {
	return job.Task.Execute(ctx)
}
