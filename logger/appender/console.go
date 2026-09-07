package appender

import (
	"io"
	"os"
	"sync"

	"github.com/anunay/logger"
	"github.com/anunay/logger/formatter"
)

// ConsoleAppender writes log entries to standard output or standard error.
type ConsoleAppender struct {
	formatter   formatter.Formatter
	out         io.Writer
	errOut      io.Writer
	splitStderr bool
	mu          sync.Mutex
}

// NewConsoleAppender creates a new ConsoleAppender with the given formatter.
func NewConsoleAppender(f formatter.Formatter) *ConsoleAppender {
	if f == nil {
		f = formatter.NewTextFormatter(true)
	}
	return &ConsoleAppender{
		formatter:   f,
		out:         os.Stdout,
		errOut:      os.Stderr,
		splitStderr: true,
	}
}

// Append writes the formatted log entry to the destination stream.
func (a *ConsoleAppender) Append(entry *logger.Entry) error {
	data, err := a.formatter.Format(entry)
	if err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	var w io.Writer = a.out
	if a.splitStderr && (entry.Level >= logger.ErrorLevel) {
		w = a.errOut
	}

	_, err = w.Write(data)
	return err
}

// Close is a no-op for standard streams.
func (a *ConsoleAppender) Close() error {
	return nil
}
