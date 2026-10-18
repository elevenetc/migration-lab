package runtime

import (
	"context"
	"fmt"
	"slices"
	"time"

	"migration-lab/backend/internal/models"
	"migration-lab/backend/internal/parser"

	"github.com/jackc/pgx/v5"
)

// DefaultDeadline stands for the grace period a pod gets before it is killed.
const DefaultDeadline = 5 * time.Second

// DefaultRows is how big every touched table is seeded, standing in for a
// production-sized table. Callers lower it to keep a run quick.
const DefaultRows int64 = 1_000_000

// Request is one runtime analysis: the migration timeline that builds the schema
// up, which of its migrations to measure, how big to seed, and how long a
// migration is allowed to take before it counts as killed.
type Request struct {
	Migrations []models.MigrationInfo
	// TargetMigrationID is the only migration to measure, its predecessors applied without
	// measurement; every migration is measured when empty.
	TargetMigrationID string
	// Rows every touched table is topped up to; DefaultRows when zero.
	Rows     int64
	Deadline time.Duration
}

// Analyse executes migrations of a timeline in one container, topping the tables
// each touches up to production-like size first, and reports how long each took,
// which locks it held, and what a killed attempt would leave behind.
//
// Measuring every migration keeps a completed one rather than rolling it back, so
// each is measured against what its predecessors left. One that does not complete
// is forced through without a deadline so the run can go on; one that fails even
// then stops it, and the migrations after it are not measured.
//
// It returns models.ErrNotFound when the timeline is empty or holds no such
// target. A failure of the run itself is part of the result rather than an error,
// since that is the answer the caller asked for.
func Analyse(ctx context.Context, request Request) (models.RuntimeAnalysisResult, error) {
	prior, toMeasure, err := split(request.Migrations, request.TargetMigrationID)
	if err != nil {
		return models.RuntimeAnalysisResult{}, err
	}
	// The timeline is parsed from the start, not just the measured migrations: an
	// operation such as ALTER COLUMN TYPE is only classifiable against the schema
	// its predecessors built.
	timeline, err := parser.ParseTimeline(slices.Concat(prior, toMeasure))
	if err != nil {
		return models.RuntimeAnalysisResult{}, err
	}
	parsed := timeline[len(prior):]

	rows, deadline := rowsOf(request.Rows), deadlineOf(request.Deadline)
	result := models.RuntimeAnalysisResult{
		DeadlineMs: deadline.Milliseconds(),
		Rows:       rows,
		Migrations: []models.RuntimeMigrationResult{},
	}

	conn, sampler, stop, err := startSession(ctx)
	if err != nil {
		result.Message = err.Error()
		return result, nil
	}
	defer stop()

	if err := applyPrior(ctx, conn, prior); err != nil {
		result.Message = err.Error()
		return result, nil
	}

	for index, info := range toMeasure {
		// The last migration is rolled back: nothing is measured after it, and a
		// rolled-back database is what a retry would find.
		last := index == len(toMeasure)-1
		migration := measureMigration(ctx, conn, sampler, info, parsed[index], rows, deadline, !last)
		result.Migrations = append(result.Migrations, migration)
		if migration.Verdict == models.RuntimeCompleted || last {
			continue
		}

		if err := finishMigration(ctx, conn, info, migration); err != nil {
			result.Stopped = info.ID
			result.Message = fmt.Sprintf("stopped at %s: %v", info.ID, err)
			return result, nil
		}
	}

	result.Message = analysisMessage(result.Migrations)
	return result, nil
}

// verdictOfRun is the verdict of the statement that ended the run; every earlier
// statement completed, or the run would have stopped there.
func verdictOfRun(measurements []models.StatementMeasurement) models.RuntimeVerdict {
	if len(measurements) == 0 {
		return models.RuntimeCompleted
	}
	return measurements[len(measurements)-1].Verdict
}

// applyPrior brings the container to the schema the analyzed migration arrives
// at, by running every migration before it.
func applyPrior(ctx context.Context, conn *pgx.Conn, migrations []models.MigrationInfo) error {
	for _, migration := range migrations {
		if _, err := conn.Exec(ctx, migration.SQL); err != nil {
			return fmt.Errorf("failed to apply %s: %w", migration.ID, err)
		}
	}
	return nil
}

func rowsOf(rows int64) int64 {
	if rows <= 0 {
		return DefaultRows
	}
	return rows
}

func deadlineOf(deadline time.Duration) time.Duration {
	if deadline <= 0 {
		return DefaultDeadline
	}
	return deadline
}

// unmeasured is a result before any measurement. The collections start empty
// rather than nil so every field of the response is an array, whichever step
// the run stops at.
func unmeasured(info models.MigrationInfo, migration *models.Migration, deadline time.Duration) models.RuntimeMigrationResult {
	return models.RuntimeMigrationResult{
		MigrationID: info.ID,
		Version:     migration.Version,
		DeadlineMs:  deadline.Milliseconds(),
		Seeded:      []models.SeededTable{},
		Statements:  []models.StatementMeasurement{},
		Findings:    []models.RuntimeFinding{},
	}
}

// split is the migrations applied unmeasured to build the schema up, and the
// ones measured after them: the named one alone, or the whole timeline when the
// caller names none. It returns models.ErrNotFound for an empty timeline or a
// name it does not hold.
func split(migrations []models.MigrationInfo, targetMigrationID string) (prior, toMeasure []models.MigrationInfo, err error) {
	if len(migrations) == 0 {
		return nil, nil, models.ErrNotFound
	}
	if targetMigrationID == "" {
		return nil, migrations, nil
	}
	for i, migration := range migrations {
		if migration.ID == targetMigrationID {
			return migrations[:i], migrations[i : i+1], nil
		}
	}
	return nil, nil, models.ErrNotFound
}

func failed(result models.RuntimeMigrationResult, message string) models.RuntimeMigrationResult {
	result.Verdict = models.RuntimeFailed
	result.Retry = models.RetryNotApplicable
	result.Message = message
	return result
}
