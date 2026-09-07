package appender

import (
	"sync"
	"time"

	"github.com/anunay/logger"
)

// AsyncAppender wraps an underlying Appender with a non-blocking asynchronous buffer and worker pool.
type AsyncAppender struct {
	underlying  Appender
	queue       chan *logger.Entry
	wg          sync.WaitGroup
	bufferSize  int
	workers     int
	blockOnFull bool
	closed      bool
	mu          sync.Mutex
}

// NewAsyncAppender wraps an existing appender with an async channel queue and background workers.
func NewAsyncAppender(underlying Appender, bufferSize, workers int) *AsyncAppender {
	if bufferSize <= 0 {
		bufferSize = 1024
	}
	if workers <= 0 {
		workers = 1
	}

	a := &AsyncAppender{
		underlying:  underlying,
		bufferSize:  bufferSize,
		workers:     workers,
		blockOnFull: false,
		queue:       make(chan *logger.Entry, bufferSize),
	}

	for i := 0; i < a.workers; i++ {
		a.wg.Add(1)
		go a.worker()
	}

	return a
}

func (a *AsyncAppender) worker() {
	defer a.wg.Done()
	for entry := range a.queue {
		_ = a.underlying.Append(entry)
	}
}

// Append puts the log entry into the async channel.
func (a *AsyncAppender) Append(entry *logger.Entry) error {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return nil
	}
	a.mu.Unlock()

	cloned := entry.Clone()

	if a.blockOnFull {
		a.queue <- cloned
		return nil
	}

	select {
	case a.queue <- cloned:
		return nil
	default:
		// Queue full & non-blocking: drop entry to avoid blocking application
		return nil
	}
}

// Flush waits until all queued logs are processed or timeout passes.
func (a *AsyncAppender) Flush(timeout time.Duration) bool {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.After(timeout)

	for {
		if len(a.queue) == 0 {
			return true
		}
		select {
		case <-deadline:
			return false
		case <-ticker.C:
		}
	}
}

// Close drains remaining queue items, shuts down workers, and closes underlying appender.
func (a *AsyncAppender) Close() error {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return nil
	}
	a.closed = true
	a.mu.Unlock()

	close(a.queue)
	a.wg.Wait()

	return a.underlying.Close()
}
