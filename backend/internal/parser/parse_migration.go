package parser

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	pgquery "github.com/pganalyze/pg_query_go/v6"
	"migration-timeline/backend/internal/models"
)

func ParseMigration(migrationID, sql string, timestamp int64) (*models.Migration, error) {
	result, err := pgquery.Parse(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQL: %w", err)
	}

	var ops []models.Operation
	for _, rawStmt := range result.Stmts {
		stmt := rawStmt.Stmt

		if createStmt := stmt.GetCreateStmt(); createStmt != nil {
			ops = append(ops, parseCreateStmt(migrationID, createStmt))
			continue
		}

		if alterStmt := stmt.GetAlterTableStmt(); alterStmt != nil {
			ops = append(ops, parseAlterTableStmt(migrationID, alterStmt)...)
			continue
		}

		if renameStmt := stmt.GetRenameStmt(); renameStmt != nil {
			if op := parseRenameStmt(migrationID, renameStmt); op != nil {
				ops = append(ops, op)
			}
			continue
		}

		if dropStmt := stmt.GetDropStmt(); dropStmt != nil {
			ops = append(ops, parseDropStmt(migrationID, dropStmt)...)
			continue
		}
	}

	version := extractVersion(migrationID)

	return &models.Migration{
		ID:         migrationID,
		Version:    version,
		Timestamp:  timestamp,
		Operations: ops,
	}, nil
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

func ExtractTimestampFromFilename(filename string) int64 {
	base := filepath.Base(filename)
	base = strings.TrimSuffix(base, filepath.Ext(base))

	re := regexp.MustCompile(`^V?(\d+)`)
	matches := re.FindStringSubmatch(base)
	if len(matches) > 1 {
		if ts, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
			return ts
		}
	}
	return 0
}
