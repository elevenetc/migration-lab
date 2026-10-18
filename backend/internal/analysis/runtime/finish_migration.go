package runtime

import (
	"context"
	"fmt"

	"migration-lab/backend/internal/models"
	"migration-lab/backend/internal/parser"

	"github.com/jackc/pgx/v5"
)

const invalidIndexNamesQuery = `
SELECT format('%I.%I', n.nspname, c.relname)
FROM pg_index i
JOIN pg_class c ON c.oid = i.indexrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE NOT i.indisvalid`

// finishMigration finishes a measured migration that did not complete under the
// deadline, the way an operator unblocks a deployment: it drops the invalid
// indexes a cancelled CONCURRENTLY build left behind, then runs the rest of the
// migration without a deadline, so the migrations after it are measured against
// the state it leaves.
func finishMigration(ctx context.Context, conn *pgx.Conn, info models.MigrationInfo, measured models.RuntimeMigrationResult) error {
	migrationStatements, err := parser.SplitStatements(info.SQL)
	if err != nil {
		return err
	}
	if _, err := conn.Exec(ctx, "SET statement_timeout = 0"); err != nil {
		return err
	}
	if err := dropInvalidIndexes(ctx, conn); err != nil {
		return err
	}

	for index := resumeIndex(migrationStatements, measured); index < len(migrationStatements); index++ {
		if _, err := conn.Exec(ctx, migrationStatements[index]); err != nil {
			return fmt.Errorf("statement %d failed even without a deadline: %w", index, err)
		}
	}
	return nil
}

// resumeIndex is the statement finishMigration restarts at. A transactional migration
// was rolled back and restarts from its first statement; any other resumes at the
// statement that stopped it, since the ones before it are already applied.
func resumeIndex(statements []string, measured models.RuntimeMigrationResult) int {
	if transactional(statements) || len(measured.Statements) == 0 {
		return 0
	}
	return measured.Statements[len(measured.Statements)-1].StatementIndex
}

func dropInvalidIndexes(ctx context.Context, conn *pgx.Conn) error {
	rows, err := conn.Query(ctx, invalidIndexNamesQuery)
	if err != nil {
		return err
	}
	names, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return err
	}
	for _, name := range names {
		if _, err := conn.Exec(ctx, "DROP INDEX "+name); err != nil {
			return fmt.Errorf("failed to drop invalid index %s: %w", name, err)
		}
	}
	return nil
}
