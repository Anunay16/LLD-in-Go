package appender

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/anunay/logger"
	"github.com/anunay/logger/formatter"
)

// FileAppender writes formatted log entries to a file.
type FileAppender struct {
	formatter formatter.Formatter
	writer    io.WriteCloser
	path      string
	mu        sync.Mutex
}

// NewFileAppender creates a FileAppender writing to the given file path.
func NewFileAppender(filePath string, f formatter.Formatter) (*FileAppender, error) {
	if f == nil {
		f = formatter.NewJSONFormatter(false)
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log dir %s: %w", dir, err)
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", filePath, err)
	}

	return &FileAppender{
		formatter: f,
		writer:    file,
		path:      filePath,
	}, nil
}

// NewCustomWriterAppender allows wrapping any custom io.WriteCloser (or io.Writer).
func NewCustomWriterAppender(w io.WriteCloser, f formatter.Formatter) *FileAppender {
	if f == nil {
		f = formatter.NewTextFormatter(false)
	}
	return &FileAppender{
		formatter: f,
		writer:    w,
	}
}

// Append formats and writes the entry to the file.
func (a *FileAppender) Append(entry *logger.Entry) error {
	data, err := a.formatter.Format(entry)
	if err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.writer == nil {
		return fmt.Errorf("file appender is closed")
	}

	_, err = a.writer.Write(data)
	return err
}

// Close closes the underlying log file.
func (a *FileAppender) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.writer != nil {
		err := a.writer.Close()
		a.writer = nil
		return err
	}
	return nil
}
