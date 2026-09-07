package formatter

import (
	"encoding/json"
	"time"

	"github.com/anunay/logger"
)

// JSONFormatter formats log entries into JSON format.
type JSONFormatter struct {
	TimestampFormat string
	PrettyPrint     bool
}

// NewJSONFormatter creates a new JSONFormatter.
func NewJSONFormatter(prettyPrint bool) *JSONFormatter {
	return &JSONFormatter{
		TimestampFormat: time.RFC3339,
		PrettyPrint:     prettyPrint,
	}
}

type jsonLogPayload struct {
	Timestamp string         `json:"timestamp"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
}

// Format transforms a LogEntry into JSON bytes.
func (f *JSONFormatter) Format(entry *logger.Entry) ([]byte, error) {
	tsFormat := f.TimestampFormat
	if tsFormat == "" {
		tsFormat = time.RFC3339
	}

	payload := jsonLogPayload{
		Timestamp: entry.Timestamp.Format(tsFormat),
		Level:     entry.Level.String(),
		Message:   entry.Message,
		Fields:    entry.Fields,
	}

	var data []byte
	var err error
	if f.PrettyPrint {
		data, err = json.MarshalIndent(payload, "", "  ")
	} else {
		data, err = json.Marshal(payload)
	}

	if err != nil {
		return nil, err
	}

	return append(data, '\n'), nil
}
