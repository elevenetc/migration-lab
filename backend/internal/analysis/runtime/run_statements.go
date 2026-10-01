package runtime

import (
	"context"
	"time"

	"migration-lab/backend/internal/models"
	"migration-lab/backend/internal/monitoring"

	"github.com/jackc/pgx/v5"
)

// runStatements measures every statement of the migration against one shared
// deadline — the grace period a pod gets — and stops at the first statement that
// does not complete, the way a killed pod does. Wrapped in a transaction it
// rolls back when the migration allows one, so the seeded database is reusable,
// unless commit asks to keep a migration that completed, so the next migration
// is measured against it without running it twice.
func runStatements(
	ctx context.Context,
	conn,
	sampler *pgx.Conn,
	migrationStatements []string,
	deadline time.Duration,
	commit bool,
) []models.StatementMeasurement {
	wrap := transactional(migrationStatements)
	if wrap {
		if _, err := conn.Exec(ctx, "BEGIN"); err != nil {
			return []models.StatementMeasurement{{
				Locks:              []models.LockObservation{},
				RewrittenRelations: []string{},
				Verdict:            models.RuntimeFailed,
				Error:              err.Error(),
			}}
		}
	}

	measurements := make([]models.StatementMeasurement, 0, len(migrationStatements))
	remaining := deadline

	for index, statement := range migrationStatements {
		measurement := measureStatement(ctx, conn, sampler, index, statement, remaining, wrap)
		measurements = append(measurements, measurement)
		remaining -= time.Duration(measurement.DurationMs) * time.Millisecond

		if measurement.Verdict != models.RuntimeCompleted {
			break
		}
	}

	// Rolling back also drops the per-transaction statement_timeout; without a
	// transaction it is a session setting the later checks must not inherit.
	if wrap {
		end := "ROLLBACK"
		if commit && verdictOfRun(measurements) == models.RuntimeCompleted {
			end = "COMMIT"
		}
		if _, err := conn.Exec(context.WithoutCancel(ctx), end); err != nil {
			monitoring.Errorf(ctx, "Failed to %s the measured migration: %v", end, err)
		}
	} else if _, err := conn.Exec(context.WithoutCancel(ctx), "SET statement_timeout = 0"); err != nil {
		monitoring.Errorf(ctx, "Failed to clear statement_timeout: %v", err)
	}

	return measurements
}
