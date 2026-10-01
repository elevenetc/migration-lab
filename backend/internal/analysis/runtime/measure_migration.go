package runtime

import (
	"context"
	"time"

	"migration-lab/backend/internal/models"
	"migration-lab/backend/internal/monitoring"
	"migration-lab/backend/internal/parser"

	"github.com/jackc/pgx/v5"
)

// measureMigration seeds the tables the migration touches in the database as its
// predecessors left it, then executes the migration under the deadline and
// reports what it measured. commit keeps a migration that completed rather than
// rolling it back.
func measureMigration(
	ctx context.Context,
	conn, sampler *pgx.Conn,
	info models.MigrationInfo,
	migration *models.Migration,
	rows int64,
	deadline time.Duration,
	commit bool,
) models.RuntimeMigrationResult {
	result := unmeasured(info, migration, deadline)

	statements, err := parser.SplitStatements(info.SQL)
	if err != nil {
		return failed(result, err.Error())
	}

	catalog, err := ReadCatalog(ctx, conn)
	if err != nil {
		monitoring.CaptureError(ctx, err)
		return failed(result, err.Error())
	}
	result.Seeded = Seed(ctx, conn, SeedPlans(catalog, TouchedTables(migration)), rows)

	monitoring.Infof(ctx, "Measuring %s: %d statements, %d ms deadline", info.ID, len(statements), deadline.Milliseconds())
	measurements := runStatements(ctx, conn, sampler, statements, deadline, commit)
	result.Statements = withClasses(migration, result.Seeded, measurements)

	result.Verdict = verdictOfRun(result.Statements)
	result.Retry = retryVerdict(ctx, conn, result.Verdict)
	result.Findings = Findings(migration, result)
	result.Message = message(result)

	return result
}
