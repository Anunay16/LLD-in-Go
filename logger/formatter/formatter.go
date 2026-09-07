package formatter

import "github.com/anunay/logger"

// Formatter defines the Strategy interface for transforming a LogEntry into formatted bytes.
type Formatter interface {
	Format(entry *logger.Entry) ([]byte, error)
}
