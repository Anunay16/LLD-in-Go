package models

import "time"

// Message represents a single piece of data in the pub-sub system.
type Message struct {
	Key       []byte
	Value     []byte
	Offset    uint64
	Timestamp time.Time
}
