# Distributed Job Scheduler in Go (LLD)

A clean, minimalist, and SOLID-compliant Low-Level Design (LLD) implementation of a **Job Scheduler** in Golang, tailored for SDE-2 / Senior Backend Engineering interviews.

---

## 📌 Architecture Diagram

```
                    ┌─────────────────────────┐
                    │        Scheduler        │  (Coordinates store, lock, executor)
                    └────────────┬────────────┘
                                 │
            ┌────────────────────┼────────────────────┐
            │                    │                    │
            ▼                    ▼                    ▼
  ┌──────────────────┐ ┌───────────────────┐ ┌──────────────────┐
  │     JobStore     │ │     Locker        │ │     Executor     │
  │  (Persistence)   │ │  (Redis/MemLock)  │ │ (Task Runner)    │
  └──────────────────┘ └─────────┬─────────┘ └────────┬─────────┘
                                 │                    │ executes
                                 ▼                    ▼
                              Redis          ┌──────────────────┐
                                             │       Job        │
                                             └────────┬─────────┘
                 ┌────────────────────────────────────┼────────────────────────────────────┐
                 │                                    │                                    │
                 ▼                                    ▼                                    ▼
    ┌─────────────────────────┐          ┌─────────────────────────┐          ┌─────────────────────────┐
    │    ScheduleStrategy     │          │      TaskStrategy       │          │   RetryPolicyStrategy   │
    ├─────────────────────────┤          ├─────────────────────────┤          ├─────────────────────────┤
    │ • OneTime               │          │ • EmailTask             │          │ • NoRetryPolicy         │
    │ • FixedInterval         │          │ • ReportTask            │          │ • ExponentialBackoff    │
    │ • CronSchedule          │          │ • NotificationTask      │          │                         │
    └─────────────────────────┘          └─────────────────────────┘          └─────────────────────────┘
```

---

## 🧱 Key Interfaces & Methods

### 1. Scheduler
```go
type Scheduler struct {
    store        store.JobStore
    lock         lock.Locker
    executor     executor.Executor
    pollInterval time.Duration
}

func (s *Scheduler) AddJob(ctx context.Context, job *model.Job) error
func (s *Scheduler) Start(ctx context.Context)
func (s *Scheduler) Stop()
```

### 2. Core Dependencies
- **`JobStore`**: `Create`, `Get`, `Update`, `Delete`, `GetDueJobs`
- **`Locker`**: `Acquire(ctx, key, ttl)`, `Release(ctx, key)`
- **`Executor`**: `Execute(ctx, job)`

### 3. Strategies
- **`ScheduleStrategy`**: `NextRun(from time.Time) *time.Time`
- **`TaskStrategy`**: `Execute(ctx context.Context) error`
- **`RetryPolicyStrategy`**: `NextRetry(attempt int) (delay time.Duration, shouldRetry bool)`

---

## 🚀 Quickstart

```bash
# Run demo
go run main.go
```
