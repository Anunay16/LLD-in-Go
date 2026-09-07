package logger

import (
	"time"
)

// Entry represents an immutable snapshot of a single log event.
type Entry struct {
	Timestamp time.Time      `json:"timestamp"`
	Level     Level          `json:"level"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
}

// Clone creates a shallow clone of the entry, copying the fields map.
func (e *Entry) Clone() *Entry {
	cloned := &Entry{
		Timestamp: e.Timestamp,
		Level:     e.Level,
		Message:   e.Message,
	}
	if len(e.Fields) > 0 {
		cloned.Fields = make(map[string]any, len(e.Fields))
		for k, v := range e.Fields {
			cloned.Fields[k] = v
		}
	}
	return cloned
}
