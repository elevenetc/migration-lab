package parser

import (
	"fmt"

	pgquery "github.com/pganalyze/pg_query_go/v6"
)

// SplitStatements returns every statement of a migration in source order,
// including the ones ParseMigration does not model (backfill UPDATEs, index
// builds). Splitting goes through pg_query, so semicolons inside literals and
// dollar-quoted bodies stay where they belong.
func SplitStatements(sql string) ([]string, error) {
	result, err := pgquery.Parse(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQL: %w", err)
	}

	statements := make([]string, 0, len(result.Stmts))
	for _, rawStmt := range result.Stmts {
		if text := statementSQL(sql, rawStmt); text != "" {
			statements = append(statements, text)
		}
	}
	return statements, nil
}
