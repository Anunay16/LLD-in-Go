package logger

import (
	"os"
	"sync"
	"time"
)

// Appender defines the sink interface for receiving log entries.
type Appender interface {
	Append(entry *Entry) error
	Close() error
}

// Logger is the core structured logger implementation.
type Logger struct {
	level    Level
	appender Appender
	fields   map[string]any
	mu       sync.RWMutex
}

// LoggerBuilder constructs a Logger instance step-by-step (Builder Pattern).
type LoggerBuilder struct {
	level    Level
	appender Appender
	fields   map[string]any
}

// NewBuilder initializes a new LoggerBuilder with default configurations.
func NewBuilder() *LoggerBuilder {
	return &LoggerBuilder{
		level:  InfoLevel,
		fields: make(map[string]any),
	}
}

// SetLevel sets the minimum logging level.
func (b *LoggerBuilder) SetLevel(lvl Level) *LoggerBuilder {
	b.level = lvl
	return b
}

// SetAppender sets the appender (sink) for log output.
func (b *LoggerBuilder) SetAppender(app Appender) *LoggerBuilder {
	b.appender = app
	return b
}

// SetDefaultFields adds default metadata fields to every log emitted.
func (b *LoggerBuilder) SetDefaultFields(fields map[string]any) *LoggerBuilder {
	for k, v := range fields {
		b.fields[k] = v
	}
	return b
}

// Build creates and returns the configured Logger instance.
func (b *LoggerBuilder) Build() *Logger {
	copiedFields := make(map[string]any, len(b.fields))
	for k, v := range b.fields {
		copiedFields[k] = v
	}
	return &Logger{
		level:    b.level,
		appender: b.appender,
		fields:   copiedFields,
	}
}

// Clone creates a child logger inheriting configuration and metadata (Prototype Pattern).
func (l *Logger) Clone() *Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	copiedFields := make(map[string]any, len(l.fields))
	for k, v := range l.fields {
		copiedFields[k] = v
	}
	return &Logger{
		level:    l.level,
		appender: l.appender,
		fields:   copiedFields,
	}
}

// With returns a child logger with additional structured metadata fields.
func (l *Logger) With(fields ...Field) *Logger {
	child := l.Clone()
	for _, f := range fields {
		child.fields[f.Key] = f.Value
	}
	return child
}

// Log emits a log entry with the specified level, message, and metadata fields.
func (l *Logger) Log(lvl Level, msg string, fields ...Field) {
	l.mu.RLock()
	if !l.level.Enabled(lvl) || l.appender == nil {
		l.mu.RUnlock()
		return
	}
	app := l.appender

	allFields := make(map[string]any, len(l.fields)+len(fields))
	for k, v := range l.fields {
		allFields[k] = v
	}
	for _, f := range fields {
		allFields[f.Key] = f.Value
	}
	l.mu.RUnlock()

	entry := &Entry{
		Timestamp: time.Now(),
		Level:     lvl,
		Message:   msg,
		Fields:    allFields,
	}

	_ = app.Append(entry)

	if lvl == FatalLevel {
		_ = app.Close()
		os.Exit(1)
	}
}

// 5 Core Level Logging Methods
func (l *Logger) Debug(msg string, fields ...Field) { l.Log(DebugLevel, msg, fields...) }
func (l *Logger) Info(msg string, fields ...Field)  { l.Log(InfoLevel, msg, fields...) }
func (l *Logger) Warn(msg string, fields ...Field)  { l.Log(WarnLevel, msg, fields...) }
func (l *Logger) Error(msg string, fields ...Field) { l.Log(ErrorLevel, msg, fields...) }
func (l *Logger) Fatal(msg string, fields ...Field) { l.Log(FatalLevel, msg, fields...) }

// Close gracefully closes the logger appender.
func (l *Logger) Close() error {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.appender != nil {
		return l.appender.Close()
	}
	return nil
}
