package parser

import (
	"fmt"

	pgquery "github.com/pganalyze/pg_query_go/v6"
)

// SplitStatements returns every statement of a migration in source order
//
// For example:
//
//	CREATE TABLE users (id INT, note TEXT DEFAULT 'a;b');
//	UPDATE users SET note = 'x';
//	CREATE INDEX idx_users_note ON users (note);
//
// becomes
//
//	[]string{
//		"CREATE TABLE users (id INT, note TEXT DEFAULT 'a;b')",
//		"UPDATE users SET note = 'x'",
//		"CREATE INDEX idx_users_note ON users (note)",
//	}
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
