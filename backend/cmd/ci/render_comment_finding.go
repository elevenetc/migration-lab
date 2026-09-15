package main

import (
	"fmt"
	"strings"

	"migration-lab/backend/internal/models"
)

// The finding's operation ID links to the measured SQL even when static
// analysis did not model that statement. Do not treat its index as a slice offset.
func renderCommentFinding(finding models.RuntimeFinding, result *models.RuntimeAnalysisResult) string {
	var sql string
	id := finding.OperationID
	if id.MigrationID == result.MigrationID && id.StatementIndex >= 0 {
		for _, statement := range result.Statements {
			if statement.StatementIndex == id.StatementIndex {
				sql = strings.TrimSpace(statement.SQL)
				break
			}
		}
	}

	message := finding.Message
	if sql != "" {
		if rest, ok := strings.CutPrefix(message, fmt.Sprintf("statement %d ", id.StatementIndex)); ok {
			message = "This statement " + rest
		}
	}
	body := "- " + commentCode(finding.Type) + ": " + commentText(message) + "\n"
	if sql != "" {
		body += commentSQL(sql)
	}
	return body
}
