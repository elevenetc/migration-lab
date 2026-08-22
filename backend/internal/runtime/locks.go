package runtime

import (
	"context"

	"github.com/jackc/pgx/v5"
	"migration-timeline/backend/internal/models"
)

// lockRank orders PostgreSQL lock modes weakest to strongest, so the strongest
// mode a statement held can be named.
var lockRank = map[string]int{
	"AccessShareLock":          1,
	"RowShareLock":             2,
	"RowExclusiveLock":         3,
	"ShareUpdateExclusiveLock": 4,
	"ShareLock":                5,
	"ShareRowExclusiveLock":    6,
	"ExclusiveLock":            7,
	"AccessExclusiveLock":      8,
}

// BlockingLock is the weakest mode that already blocks plain readers.
const BlockingLock = "AccessExclusiveLock"

// The join drops locks on relations that no longer exist in pg_class — the
// transient heap a table rewrite builds and swaps in — leaving the relations a
// reader of this schema can actually be blocked on.
const heldLocksQuery = `
SELECT l.mode, c.relname
FROM pg_locks l
JOIN pg_class c ON c.oid = l.relation
WHERE l.pid = $1 AND l.granted AND l.locktype = 'relation'`

// Backends waiting on a lock the migration holds. PostgreSQL answers this
// exactly, so no latency threshold has to stand in for "blocked".
const blockedReadersQuery = `
SELECT pid
FROM pg_stat_activity
WHERE wait_event_type = 'Lock' AND $1 = ANY(pg_blocking_pids(pid))`

// StrongestLock is the strongest mode among the observations, "" for none.
func StrongestLock(locks []models.LockObservation) string {
	strongest := ""
	for _, lock := range locks {
		if lockRank[lock.Mode] > lockRank[strongest] {
			strongest = lock.Mode
		}
	}
	return strongest
}

// BlocksReaders reports whether the mode keeps a plain SELECT waiting.
func BlocksReaders(mode string) bool {
	return lockRank[mode] >= lockRank[BlockingLock]
}

// sampleLocks reads the relation locks a backend currently holds. It runs on a
// connection of its own, since the backend being sampled is busy.
func sampleLocks(ctx context.Context, conn *pgx.Conn, pid uint32) ([]models.LockObservation, error) {
	rows, err := conn.Query(ctx, heldLocksQuery, pid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locks []models.LockObservation
	for rows.Next() {
		var lock models.LockObservation
		if err := rows.Scan(&lock.Mode, &lock.Relation); err != nil {
			return nil, err
		}
		locks = append(locks, lock)
	}
	return locks, rows.Err()
}

// sampleBlockedReaders reads which backends are waiting on a lock the migration's
// backend holds, right now.
func sampleBlockedReaders(ctx context.Context, conn *pgx.Conn, pid uint32) ([]uint32, error) {
	rows, err := conn.Query(ctx, blockedReadersQuery, pid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocked []uint32
	for rows.Next() {
		var blockedPID uint32
		if err := rows.Scan(&blockedPID); err != nil {
			return nil, err
		}
		blocked = append(blocked, blockedPID)
	}
	return blocked, rows.Err()
}
