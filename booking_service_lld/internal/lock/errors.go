package lock

import "errors"

var (
	ErrLockNotAcquired        = errors.New("could not acquire lock")
	ErrLockNotOwnedOrReleased = errors.New("lock not owned or already released")
)
