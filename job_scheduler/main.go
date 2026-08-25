package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"job_scheduler/pkg/executor"
	"job_scheduler/pkg/lock"
	"job_scheduler/pkg/model"
	"job_scheduler/pkg/scheduler"
	"job_scheduler/pkg/store"
	"job_scheduler/pkg/strategy"
)

// FlakyTask demonstrates retry mechanism by failing twice before succeeding.
type FlakyTask struct {
	attempts int
}

func (f *FlakyTask) Execute(ctx context.Context) error {
	f.attempts++
	if f.attempts < 3 {
		return fmt.Errorf("simulated network timeout (attempt %d)", f.attempts)
	}
	fmt.Printf("[FlakyTask] Successfully executed on attempt %d!\n", f.attempts)
	return nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Initialize Components (Directly matching Architecture Diagram)
	jobStore := store.NewMemoryJobStore()
	redisLocker := lock.NewRedisLock("redis://localhost:6379")
	jobExecutor := executor.NewDefaultExecutor()

	// Scheduler directly wires JobStore, Lock, and Executor
	sched := scheduler.NewScheduler(jobStore, redisLocker, jobExecutor, 500*time.Millisecond)

	// 2. Define Jobs with different Strategy combinations

	// Job 1: One-time Email Job (triggers immediately)
	job1 := model.NewJob(
		"job-1",
		"Welcome Email",
		strategy.NewOneTimeSchedule(time.Now()),
		strategy.NewEmailTask("user@example.com", "Welcome to Platform!", "Thanks for joining us."),
		strategy.NewNoRetryPolicy(),
	)

	// Job 2: Fixed Interval Slack Notification (every 2 seconds)
	job2 := model.NewJob(
		"job-2",
		"Slack Health Ping",
		strategy.NewFixedIntervalSchedule(2*time.Second),
		strategy.NewNotificationTask("#devops", "System health check: OK"),
		strategy.NewExponentialBackoffPolicy(3, 1*time.Second, 5*time.Second),
	)

	// Job 3: Recurring Cron Report Generation (every 3 seconds)
	job3 := model.NewJob(
		"job-3",
		"Daily Sales Report",
		strategy.NewCronSchedule("0 0 * * *", 3*time.Second),
		strategy.NewReportTask("Daily Sales", "PDF"),
		strategy.NewExponentialBackoffPolicy(2, 500*time.Millisecond, 2*time.Second),
	)

	// Job 4: Flaky Job to showcase Exponential Backoff Retry Policy
	job4 := model.NewJob(
		"job-4",
		"Flaky Data Sync",
		strategy.NewOneTimeSchedule(time.Now()),
		&FlakyTask{},
		strategy.NewExponentialBackoffPolicy(3, 500*time.Millisecond, 3*time.Second),
	)

	// Register jobs
	_ = sched.AddJob(ctx, job1)
	_ = sched.AddJob(ctx, job2)
	_ = sched.AddJob(ctx, job3)
	_ = sched.AddJob(ctx, job4)

	fmt.Println("==================================================")
	fmt.Println("🚀 Starting Distributed Job Scheduler Demo...")
	fmt.Println("==================================================")

	sched.Start(ctx)

	// Run for 8 seconds or until interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigChan:
		fmt.Println("\nInterrupt signal received. Shutting down...")
	case <-time.After(8 * time.Second):
		fmt.Println("\nDemo completed. Shutting down...")
	}

	sched.Stop()
	fmt.Println("✅ Scheduler stopped gracefully.")
}
