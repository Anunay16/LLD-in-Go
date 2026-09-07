package appender

import "github.com/anunay/logger"

// Appender represents the Observer / Sink interface for consuming log entries.
type Appender interface {
	Append(entry *logger.Entry) error
	Close() error
}
