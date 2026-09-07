package appender

import (
	"errors"
	"sync"

	"github.com/anunay/logger"
)

// MultiAppender forwards log entries to multiple underlying appenders (Composite Pattern).
type MultiAppender struct {
	appenders []Appender
	mu        sync.RWMutex
}

// NewMultiAppender creates a composite MultiAppender.
func NewMultiAppender(appenders ...Appender) *MultiAppender {
	return &MultiAppender{
		appenders: appenders,
	}
}

// AddAppender dynamically adds a new appender to the composite.
func (m *MultiAppender) AddAppender(a Appender) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.appenders = append(m.appenders, a)
}

// Append sends the log entry to all child appenders.
func (m *MultiAppender) Append(entry *logger.Entry) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errs []error
	for _, app := range m.appenders {
		if err := app.Append(entry); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Close closes all underlying child appenders.
func (m *MultiAppender) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for _, app := range m.appenders {
		if err := app.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
