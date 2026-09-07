package logger

import (
	"time"
)

// Field represents a key-value pair for structured logging metadata.
type Field struct {
	Key   string
	Value any
}

// Helper constructor functions for Field

func String(key, val string) Field {
	return Field{Key: key, Value: val}
}

func Int(key string, val int) Field {
	return Field{Key: key, Value: val}
}

func Int64(key string, val int64) Field {
	return Field{Key: key, Value: val}
}

func Float64(key string, val float64) Field {
	return Field{Key: key, Value: val}
}

func Bool(key string, val bool) Field {
	return Field{Key: key, Value: val}
}

func Err(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}

func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Value: val.String()}
}

func Time(key string, val time.Time) Field {
	return Field{Key: key, Value: val}
}

func Any(key string, val any) Field {
	return Field{Key: key, Value: val}
}
