package strategy

import "time"

// ScheduleStrategy defines the contract for computing the next execution time.
type ScheduleStrategy interface {
	NextRun(from time.Time) *time.Time
}

// OneTimeSchedule triggers once at a specified time.
type OneTimeSchedule struct {
	ExecutionTime time.Time
	triggered     bool
}

// NewOneTimeSchedule creates a schedule for a single point in time.
func NewOneTimeSchedule(execTime time.Time) *OneTimeSchedule {
	return &OneTimeSchedule{
		ExecutionTime: execTime,
		triggered:     false,
	}
}

func (s *OneTimeSchedule) NextRun(from time.Time) *time.Time {
	if s.triggered {
		return nil
	}
	s.triggered = true
	t := s.ExecutionTime
	return &t
}

// FixedIntervalSchedule triggers periodically at fixed intervals.
type FixedIntervalSchedule struct {
	Interval time.Duration
}

// NewFixedIntervalSchedule creates a recurring schedule with a fixed duration interval.
func NewFixedIntervalSchedule(interval time.Duration) *FixedIntervalSchedule {
	return &FixedIntervalSchedule{Interval: interval}
}

func (s *FixedIntervalSchedule) NextRun(from time.Time) *time.Time {
	next := from.Add(s.Interval)
	return &next
}

// CronSchedule simulates cron expression based scheduling.
type CronSchedule struct {
	Interval time.Duration // Simplified for LLD demo; in production use cron expr parser
	Name     string
}

// NewCronSchedule creates a cron-based schedule.
func NewCronSchedule(name string, interval time.Duration) *CronSchedule {
	return &CronSchedule{
		Name:     name,
		Interval: interval,
	}
}

func (s *CronSchedule) NextRun(from time.Time) *time.Time {
	next := from.Add(s.Interval)
	return &next
}
