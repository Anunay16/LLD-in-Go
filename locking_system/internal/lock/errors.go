package lock

import "errors"

var (
	ErrLockNotAcquired = errors.New("couldn't acquire lock")
	ErrLockNotOwnedOrReleased = errors.New("lock not owned or already released")
)
