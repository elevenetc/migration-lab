package runtime

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"migration-lab/backend/internal/models"
)

// queryCanceled is the SQLSTATE statement_timeout raises, which is how a pod
// killed part-way through a migration shows up here.
const queryCanceled = "57014"

// measureStatement runs one statement under the deadline left for it while a
// second connection samples the locks its backend holds and who waits behind
// them. The schema is snapshotted around the statement — outside the timed
// window, so the duration stays comparable — which is what the observed
// performance class is derived from. The blocked-backend tick counts are returned
// for the caller to attribute to the reader probes once the whole run is over.
func measureStatement(ctx context.Context, conn, sampler *pgx.Conn, index int, sql string, budget time.Duration, countersUsable bool) (models.StatementMeasurement, map[uint32]int) {
	measurement := models.StatementMeasurement{
		StatementIndex:     index,
		SQL:                sql,
		Locks:              []models.LockObservation{},
		RewrittenRelations: []string{},
	}

	if budget <= 0 {
		measurement.Verdict = models.RuntimeExceedsDeadline
		measurement.Error = "the deadline was spent by the statements before it"
		return measurement, nil
	}

	before, err := readRelations(ctx, conn)
	if err != nil {
		log.Printf("Failed to snapshot the schema before statement %d: %v", index, err)
	}

	if _, err := conn.Exec(ctx, fmt.Sprintf("SET statement_timeout = %d", budget.Milliseconds())); err != nil {
		measurement.Verdict = models.RuntimeFailed
		measurement.Error = err.Error()
		return measurement, nil
	}

	stopSampling := startLockSampler(ctx, sampler, conn.PgConn().PID())
	start := time.Now()
	_, err = conn.Exec(ctx, sql)
	measurement.DurationMs = time.Since(start).Milliseconds()

	sample := stopSampling()
	measurement.Locks = sample.Locks
	measurement.StrongestLock = StrongestLock(sample.Locks)
	measurement.Verdict = verdictOf(err)
	if err != nil {
		measurement.Error = err.Error()
		return measurement, sample.BlockedTicks
	}

	seen := observeAfter(ctx, conn, index, before, countersUsable)
	measurement.ObservedClass = seen.Class
	measurement.RewrittenRelations = seen.Rewritten
	measurement.TuplesRead = seen.TuplesRead

	return measurement, sample.BlockedTicks
}

// observeAfter snapshots the schema again and compares it with the snapshot taken
// before the statement. It is only reachable for a statement that completed: a
// failed or cancelled one leaves the transaction aborted, where no query runs.
func observeAfter(ctx context.Context, conn *pgx.Conn, index int, before map[uint32]relationSnapshot, countersUsable bool) observation {
	// The budget is spent; the snapshot must not inherit it and be cancelled itself.
	if _, err := conn.Exec(ctx, "SET statement_timeout = 0"); err != nil {
		log.Printf("Failed to clear statement_timeout after statement %d: %v", index, err)
		return observation{Rewritten: []string{}}
	}

	after, err := readRelations(ctx, conn)
	if err != nil {
		log.Printf("Failed to snapshot the schema after statement %d: %v", index, err)
		return observation{Rewritten: []string{}}
	}
	return observe(before, after, countersUsable)
}

func verdictOf(err error) models.RuntimeVerdict {
	if err == nil {
		return models.RuntimeCompleted
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == queryCanceled {
		return models.RuntimeExceedsDeadline
	}
	return models.RuntimeFailed
}
