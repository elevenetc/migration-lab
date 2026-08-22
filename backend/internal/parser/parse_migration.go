package parser

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"migration-timeline/backend/internal/models"

	pgquery "github.com/pganalyze/pg_query_go/v6"
)

func ParseMigration(migrationID, sql string, timestamp int64) (*models.Migration, error) {
	result, err := pgquery.Parse(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQL: %w", err)
	}

	// Empty rather than nil: a migration with no supported statement must marshal
	// as [] so it stays assignable to the frontend's Statement[]. Do not replace
	// this with `var statements []models.Statement`.
	statements := []models.Statement{}
	for _, rawStmt := range result.Stmts {
		stmt := rawStmt.Stmt

		var kind string
		var ops []models.Operation

		switch {
		case stmt.GetCreateStmt() != nil:
			kind = "CREATE_TABLE"
			ops = []models.Operation{parseCreateStmt(stmt.GetCreateStmt())}
		case stmt.GetAlterTableStmt() != nil:
			kind = "ALTER_TABLE"
			ops = parseAlterTableStmt(stmt.GetAlterTableStmt())
		case stmt.GetRenameStmt() != nil:
			kind = "RENAME"
			if op := parseRenameStmt(stmt.GetRenameStmt()); op != nil {
				ops = []models.Operation{op}
			}
		case stmt.GetDropStmt() != nil:
			kind = "DROP_TABLE"
			ops = parseDropStmt(stmt.GetDropStmt())
		}

		if len(ops) == 0 {
			continue
		}

		statements = append(statements, models.Statement{
			Index:      len(statements),
			Kind:       kind,
			SQL:        statementSQL(sql, rawStmt),
			Operations: ops,
		})
	}

	version := extractVersion(migrationID)

	return &models.Migration{
		ID:         migrationID,
		Version:    version,
		Timestamp:  timestamp,
		Statements: statements,
		SQL:        sql,
	}, nil
}

// statementSQL slices the statement text out of the migration SQL using the
// byte offsets pg_query reports; StmtLen is 0 for the last statement.
func statementSQL(sql string, rawStmt *pgquery.RawStmt) string {
	start := int(rawStmt.StmtLocation)
	if start < 0 || start > len(sql) {
		return ""
	}
	end := len(sql)
	if rawStmt.StmtLen > 0 {
		end = min(start+int(rawStmt.StmtLen), len(sql))
	}
	return strings.TrimSpace(sql[start:end])
}

func extractVersion(migrationID string) string {
	base := filepath.Base(migrationID)
	base = strings.TrimSuffix(base, filepath.Ext(base))

	re := regexp.MustCompile(`^V?(\d+(?:_\d+)*)`)
	matches := re.FindStringSubmatch(base)
	if len(matches) > 1 {
		return strings.ReplaceAll(matches[1], "_", ".")
	}

	return base
}

// CompareVersions reports whether migration a's version sorts before b's, using
// Flyway's segment-wise numeric ordering: segments are compared as numbers
// ("20260410.1" < "20260410.10", not lexicographically), and a shorter prefix
// sorts first ("1" < "1.1").
func CompareVersions(a, b string) bool {
	as := strings.Split(extractVersion(a), ".")
	bs := strings.Split(extractVersion(b), ".")
	for i := 0; i < len(as) && i < len(bs); i++ {
		if c := compareSegment(as[i], bs[i]); c != 0 {
			return c < 0
		}
	}
	return len(as) < len(bs)
}

func compareSegment(a, b string) int {
	ai, aErr := strconv.ParseInt(a, 10, 64)
	bi, bErr := strconv.ParseInt(b, 10, 64)
	if aErr != nil || bErr != nil {
		return strings.Compare(a, b)
	}
	switch {
	case ai < bi:
		return -1
	case ai > bi:
		return 1
	default:
		return 0
	}
}
