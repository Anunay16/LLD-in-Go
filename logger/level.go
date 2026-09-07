package logger

import (
	"fmt"
	"strings"
)

// Level represents severity of the log message.
type Level int8

const (
	// DebugLevel is useful for debugging in development.
	DebugLevel Level = iota
	// InfoLevel is default operational logging.
	InfoLevel
	// WarnLevel represents non-critical warnings.
	WarnLevel
	// ErrorLevel indicates an error that needs attention.
	ErrorLevel
	// FatalLevel logs critical error and terminates execution.
	FatalLevel
)

// String returns the string representation of the Level.
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return fmt.Sprintf("LEVEL(%d)", l)
	}
}

// ParseLevel parses a string representation to a Level.
func ParseLevel(lvl string) (Level, error) {
	switch strings.ToUpper(strings.TrimSpace(lvl)) {
	case "DEBUG":
		return DebugLevel, nil
	case "INFO":
		return InfoLevel, nil
	case "WARN", "WARNING":
		return WarnLevel, nil
	case "ERROR":
		return ErrorLevel, nil
	case "FATAL":
		return FatalLevel, nil
	default:
		return InfoLevel, fmt.Errorf("unknown log level: %q", lvl)
	}
}

// Enabled returns true if the current level allows the target level.
func (l Level) Enabled(target Level) bool {
	return target >= l
}
