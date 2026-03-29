package bookingservice

import (
	"context"
	"fmt"
	"sort"
)

func BookSeats(ctx context.Context, user string, seats []string) error {
	sort.Strings(seats)

	var locks []Lock

	for _, seat := range seats {
		lock, err := provider.TryAcquire(ctx, seat, user, ttl)
		if err != nil {
			// rollback
			for _, l := range locks {
				l.Unlock(ctx)
			}
			return fmt.Errorf("seat %s unavailable", seat)
		}
		locks = append(locks, lock)
	}

	// success
	return nil
}
