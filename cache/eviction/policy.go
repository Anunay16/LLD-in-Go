package eviction

import "errors"

// ErrEmptyPolicy is returned when attempting to evict from an empty eviction policy tracker.
var ErrEmptyPolicy = errors.New("eviction policy has no elements to evict")

// Policy defines the Strategy interface for cache eviction algorithms.
type Policy interface {
	// KeyAccessed notifies the strategy that an existing key was accessed.
	KeyAccessed(key string)

	// KeyAdded notifies the strategy that a new key has been inserted.
	KeyAdded(key string)

	// KeyRemoved notifies the strategy that a key was deleted.
	KeyRemoved(key string)

	// Evict selects and removes a victim key according to the strategy rules.
	Evict() (string, error)
}
