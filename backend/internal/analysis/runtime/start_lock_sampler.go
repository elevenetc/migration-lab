package runtime

import (
	"context"
	"sort"
	"time"

	"migration-timeline/backend/internal/models"

	"github.com/jackc/pgx/v5"
)

const lockSampleInterval = 20 * time.Millisecond

// lockSample is what the sampler saw while a statement ran.
type lockSample struct {
	Locks []models.LockObservation
	// BlockedTicks counts the ticks in which a backend was waiting on a lock the
	// migration held, keyed by that backend's pid. Multiplied by the sample
	// interval it is how long that backend waited.
	BlockedTicks map[uint32]int
}

// startLockSampler polls what the migration's backend holds, and who is stuck
// behind it, until the returned stop function is called. Sampling runs on a
// connection of its own, since the sampled backend is busy running DDL.
func startLockSampler(ctx context.Context, sampler *pgx.Conn, pid uint32) func() lockSample {
	stopped := make(chan struct{})
	done := make(chan lockSample, 1)

	go func() {
		seen := map[models.LockObservation]bool{}
		blockedTicks := map[uint32]int{}

		for {
			if locks, err := sampleLocks(ctx, sampler, pid); err == nil {
				for _, lock := range locks {
					seen[lock] = true
				}
			}
			if blocked, err := sampleBlockedReaders(ctx, sampler, pid); err == nil {
				for _, blockedPID := range blocked {
					blockedTicks[blockedPID]++
				}
			}

			select {
			case <-stopped:
				done <- lockSample{Locks: sortedLocks(seen), BlockedTicks: blockedTicks}
				return
			case <-time.After(lockSampleInterval):
			}
		}
	}()

	return func() lockSample {
		close(stopped)
		return <-done
	}
}

func sortedLocks(seen map[models.LockObservation]bool) []models.LockObservation {
	locks := make([]models.LockObservation, 0, len(seen))
	for lock := range seen {
		locks = append(locks, lock)
	}
	sort.Slice(locks, func(i, j int) bool {
		if locks[i].Relation != locks[j].Relation {
			return locks[i].Relation < locks[j].Relation
		}
		return locks[i].Mode < locks[j].Mode
	})
	return locks
}
