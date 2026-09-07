package formatter

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/anunay/logger"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
)

// TextFormatter formats log entries into human-readable text.
type TextFormatter struct {
	TimestampFormat string
	EnableColors    bool
}

// NewTextFormatter creates a TextFormatter instance.
func NewTextFormatter(enableColors bool) *TextFormatter {
	return &TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		EnableColors:    enableColors,
	}
}

func (f *TextFormatter) getLevelColor(level logger.Level) string {
	if !f.EnableColors {
		return ""
	}
	switch level {
	case logger.DebugLevel:
		return colorCyan
	case logger.InfoLevel:
		return colorGreen
	case logger.WarnLevel:
		return colorYellow
	case logger.ErrorLevel:
		return colorRed
	case logger.FatalLevel:
		return colorPurple
	default:
		return colorReset
	}
}

// Format transforms a LogEntry into formatted text bytes.
func (f *TextFormatter) Format(entry *logger.Entry) ([]byte, error) {
	var buf bytes.Buffer

	tsFormat := f.TimestampFormat
	if tsFormat == "" {
		tsFormat = "2006-01-02 15:04:05.000"
	}

	// [TIMESTAMP]
	buf.WriteString(entry.Timestamp.Format(tsFormat))
	buf.WriteString(" ")

	// [LEVEL] with optional color
	lvlColor := f.getLevelColor(entry.Level)
	if f.EnableColors && lvlColor != "" {
		buf.WriteString(lvlColor)
	}
	buf.WriteString(fmt.Sprintf("[% -5s]", entry.Level.String()))
	if f.EnableColors && lvlColor != "" {
		buf.WriteString(colorReset)
	}
	buf.WriteString(" ")

	// Message
	buf.WriteString(entry.Message)

	// Fields / Metadata (sorted for deterministic output)
	if len(entry.Fields) > 0 {
		buf.WriteString(" {")
		keys := make([]string, 0, len(entry.Fields))
		for k := range entry.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for i, k := range keys {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(fmt.Sprintf("%s=%v", k, entry.Fields[k]))
		}
		buf.WriteString("}")
	}

	buf.WriteString("\n")
	return buf.Bytes(), nil
}
