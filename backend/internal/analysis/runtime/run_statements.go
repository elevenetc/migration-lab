package runtime

import (
	"context"
	"log"
	"strings"
	"time"

	"migration-timeline/backend/internal/models"

	"github.com/jackc/pgx/v5"
)

// Statements PostgreSQL refuses to run inside a transaction block; a migration
// containing one cannot be measured and rolled back.
var nonTransactional = []string{
	"CONCURRENTLY", "VACUUM", "REINDEX", "ALTER SYSTEM",
	"CREATE DATABASE", "DROP DATABASE", "CREATE TABLESPACE", "DROP TABLESPACE",
}

// transactional reports whether the whole migration can run inside one
// transaction, which lets the measurement roll back and leave the seeded
// database as it found it.
func transactional(statements []string) bool {
	for _, statement := range statements {
		upper := strings.ToUpper(statement)
		for _, keyword := range nonTransactional {
			if strings.Contains(upper, keyword) {
				return false
			}
		}
	}
	return true
}

// runOutcome is the whole measured migration: what each statement did, and how
// many sampler ticks each other backend spent waiting behind it.
type runOutcome struct {
	Measurements []models.StatementMeasurement
	BlockedTicks map[uint32]int
}

// runStatements measures every statement of the migration against one shared
// deadline — the grace period a pod gets — and stops at the first statement that
// does not complete, the way a killed pod does. Wrapped in a transaction it
// rolls back when the migration allows one, so the seeded database is reusable.
func runStatements(ctx context.Context, conn, sampler *pgx.Conn, statements []string, deadline time.Duration) runOutcome {
	wrap := transactional(statements)
	if wrap {
		if _, err := conn.Exec(ctx, "BEGIN"); err != nil {
			return runOutcome{Measurements: []models.StatementMeasurement{{
				Locks:              []models.LockObservation{},
				RewrittenRelations: []string{},
				Verdict:            models.RuntimeFailed,
				Error:              err.Error(),
			}}}
		}
	}

	outcome := runOutcome{
		Measurements: make([]models.StatementMeasurement, 0, len(statements)),
		BlockedTicks: map[uint32]int{},
	}
	remaining := deadline

	for index, statement := range statements {
		measurement, blocked := measureStatement(ctx, conn, sampler, index, statement, remaining, wrap)
		outcome.Measurements = append(outcome.Measurements, measurement)
		for pid, ticks := range blocked {
			outcome.BlockedTicks[pid] += ticks
		}
		remaining -= time.Duration(measurement.DurationMs) * time.Millisecond

		if measurement.Verdict != models.RuntimeCompleted {
			break
		}
	}

	// Rolling back also drops the per-transaction statement_timeout; without a
	// transaction it is a session setting the later checks must not inherit.
	if wrap {
		if _, err := conn.Exec(context.WithoutCancel(ctx), "ROLLBACK"); err != nil {
			log.Printf("Failed to roll back the measured migration: %v", err)
		}
	} else if _, err := conn.Exec(context.WithoutCancel(ctx), "SET statement_timeout = 0"); err != nil {
		log.Printf("Failed to clear statement_timeout: %v", err)
	}

	return outcome
}
