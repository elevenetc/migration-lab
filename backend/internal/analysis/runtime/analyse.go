package runtime

import (
	"context"
	"fmt"
	"log"
	"time"

	"migration-lab/backend/internal/models"
	"migration-lab/backend/internal/parser"
	"migration-lab/backend/internal/pg"

	"github.com/jackc/pgx/v5"
)

// DefaultDeadline stands for the grace period a pod gets before it is killed.
const DefaultDeadline = 5 * time.Second

// DefaultRows is how big every touched table is seeded, standing in for a
// production-sized table. Callers lower it to keep a run quick.
const DefaultRows int64 = 1_000_000

// Request is one runtime analysis: the migration timeline that builds the schema
// up, which of its migrations to measure, how big to seed, and how long the
// migration is allowed to take before it counts as killed.
type Request struct {
	Migrations []models.MigrationInfo
	// Target is the migration to measure; the last of the timeline when empty,
	// which is the one that just arrived on a branch.
	Target string
	// Rows every touched table is seeded to; DefaultRows when zero.
	Rows     int64
	Deadline time.Duration
}

// Analyse executes one migration of a timeline against a container seeded to
// production-like size, and reports how long it took, which locks it held,
// and what a killed attempt would leave behind.
//
// It returns models.ErrNotFound when the timeline is empty or holds no such
// migration. A failure of the run itself is part of the result rather than an
// error, since that is the answer the caller asked for.
func Analyse(ctx context.Context, request Request) (models.RuntimeAnalysisResult, error) {
	index, err := targetIndex(request.Migrations, request.Target)
	if err != nil {
		return models.RuntimeAnalysisResult{}, err
	}

	info := request.Migrations[index]
	// The whole timeline is parsed, not just the analyzed migration: an operation
	// such as ALTER COLUMN TYPE is only classifiable against the schema its
	// predecessors built.
	timeline, err := parser.ParseTimeline(request.Migrations[:index+1])
	if err != nil {
		return models.RuntimeAnalysisResult{}, err
	}
	migration := timeline[index]
	statements, err := parser.SplitStatements(info.SQL)
	if err != nil {
		return models.RuntimeAnalysisResult{}, err
	}

	deadline := request.Deadline
	if deadline <= 0 {
		deadline = DefaultDeadline
	}
	// The collections start empty rather than nil so every field of the response
	// is an array, whichever step the run stops at.
	result := models.RuntimeAnalysisResult{
		MigrationID: info.ID,
		Version:     migration.Version,
		DeadlineMs:  deadline.Milliseconds(),
		Seeded:      []models.SeededTable{},
		Statements:  []models.StatementMeasurement{},
		Findings:    []models.RuntimeFinding{},
	}

	connString, terminate, err := pg.Start(ctx)
	if err != nil {
		return failed(result, err.Error()), nil
	}
	defer terminate()

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return failed(result, fmt.Sprintf("failed to connect to the database: %v", err)), nil
	}
	defer func() {
		if err := conn.Close(context.WithoutCancel(ctx)); err != nil {
			log.Printf("Failed to close runtime connection: %v", err)
		}
	}()

	if err := applyPrior(ctx, conn, request.Migrations[:index]); err != nil {
		return failed(result, err.Error()), nil
	}

	catalog, err := ReadCatalog(ctx, conn)
	if err != nil {
		return failed(result, err.Error()), nil
	}
	result.Seeded = Seed(ctx, conn, SeedPlans(catalog, TouchedTables(migration)), rowsOf(request))

	sampler, err := pgx.Connect(ctx, connString)
	if err != nil {
		return failed(result, fmt.Sprintf("failed to open a sampling connection: %v", err)), nil
	}
	defer func() {
		if err := sampler.Close(context.WithoutCancel(ctx)); err != nil {
			log.Printf("Failed to close lock sampler connection: %v", err)
		}
	}()

	log.Printf("Measuring %s: %d statements, %d ms deadline", info.ID, len(statements), deadline.Milliseconds())
	measurements := runStatements(ctx, conn, sampler, statements, deadline)
	result.Statements = withClasses(migration, result.Seeded, measurements)

	result.Verdict = verdictOfRun(result.Statements)
	result.Retry = retryVerdict(ctx, conn, result.Verdict)
	result.Findings = Findings(migration, result)
	result.Message = message(result)

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

func rowsOf(request Request) int64 {
	if request.Rows <= 0 {
		return DefaultRows
	}
	return request.Rows
}

// targetIndex is the position of the migration to measure: the named one, or the
// last of the timeline when the caller names none. It returns models.ErrNotFound
// for an empty timeline or a name it does not hold.
func targetIndex(migrations []models.MigrationInfo, target string) (int, error) {
	if len(migrations) == 0 {
		return 0, models.ErrNotFound
	}
	if target == "" {
		return len(migrations) - 1, nil
	}
	for i, migration := range migrations {
		if migration.ID == target {
			return i, nil
		}
	}
	return 0, models.ErrNotFound
}

func failed(result models.RuntimeAnalysisResult, message string) models.RuntimeAnalysisResult {
	result.Verdict = models.RuntimeFailed
	result.Retry = models.RetryNotApplicable
	result.Message = message
	return result
}
