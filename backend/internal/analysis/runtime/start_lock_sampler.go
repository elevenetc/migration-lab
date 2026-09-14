package runtime

import (
	"context"
	"sort"
	"time"

	"migration-lab/backend/internal/models"

	"github.com/jackc/pgx/v5"
)

const lockSampleInterval = 20 * time.Millisecond

// startLockSampler polls the locks the migration's backend holds until the
// returned stop function is called. Sampling runs on a connection of its own,
// since the sampled backend is busy running DDL.
func startLockSampler(ctx context.Context, sampler *pgx.Conn, pid uint32) func() []models.LockObservation {
	stopped := make(chan struct{})
	done := make(chan []models.LockObservation, 1)

	go func() {
		seen := map[models.LockObservation]bool{}

		for {
			if locks, err := sampleLocks(ctx, sampler, pid); err == nil {
				for _, lock := range locks {
					seen[lock] = true
				}
			}
			select {
			case <-stopped:
				done <- sortedLocks(seen)
				return
			case <-time.After(lockSampleInterval):
			}
		}
	}()

	return func() []models.LockObservation {
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
