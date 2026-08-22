package runtime

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"migration-timeline/backend/internal/models"
)

// queryCanceled is the SQLSTATE statement_timeout raises, which is how a pod
// killed part-way through a migration shows up here.
const queryCanceled = "57014"

// measureStatement runs one statement under the deadline left for it while a
// second connection samples the locks its backend holds and who waits behind
// them. The blocked-backend tick counts are returned for the caller to attribute
// to the reader probes once the whole run is over.
func measureStatement(ctx context.Context, conn, sampler *pgx.Conn, index int, sql string, budget time.Duration) (models.StatementMeasurement, map[uint32]int) {
	measurement := models.StatementMeasurement{
		StatementIndex: index,
		SQL:            sql,
		Locks:          []models.LockObservation{},
	}

	if budget <= 0 {
		measurement.Verdict = models.RuntimeExceedsDeadline
		measurement.Error = "the deadline was spent by the statements before it"
		return measurement, nil
	}

	if _, err := conn.Exec(ctx, fmt.Sprintf("SET statement_timeout = %d", budget.Milliseconds())); err != nil {
		measurement.Verdict = models.RuntimeFailed
		measurement.Error = err.Error()
		return measurement, nil
	}

	stopSampling := startLockSampler(ctx, sampler, conn.PgConn().PID())
	start := time.Now()
	_, err := conn.Exec(ctx, sql)
	measurement.DurationMs = time.Since(start).Milliseconds()

	sample := stopSampling()
	measurement.Locks = sample.Locks
	measurement.StrongestLock = StrongestLock(sample.Locks)
	measurement.Verdict = verdictOf(err)
	if err != nil {
		measurement.Error = err.Error()
	}
	return measurement, sample.BlockedTicks
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
