package prcomment

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"migration-lab/backend/internal/models"
)

func TestCommentShowsTheMeasuredSQLForTheFinding(t *testing.T) {
	const target = "V2__add_nonnegative_balance_check.sql"
	const sql = "ALTER TABLE accounts\n    ADD CONSTRAINT nonnegative_balance CHECK (balance >= 0);"
	result := &models.RuntimeAnalysisResult{
		MigrationID: target, Verdict: models.RuntimeCompleted,
		Statements: []models.StatementMeasurement{
			{StatementIndex: 7, SQL: "SELECT 'unrelated statement';", ObservedClass: models.MetadataOnly},
			{StatementIndex: 0, SQL: sql, ObservedClass: models.DataScanning},
		},
		Findings: []models.RuntimeFinding{{
			Type:        models.FindingExclusiveLock,
			OperationID: models.OperationID{MigrationID: target, StatementIndex: 0, OpIndex: models.StatementScoped},
			Message:     "statement 0 held AccessExclusiveLock on accounts for 54 ms, and its cost scales with table size, so the hold grows with the table",
		}},
	}
	body := RenderPRComment(Summary{Migrations: []Migration{{Name: target, Runtime: result}}}, testCommentRun())
	want := "- ` EXCLUSIVE_LOCK_HELD `: This statement held AccessExclusiveLock on accounts for 54 ms, and its cost scales with table size, so the hold grows with the table\n\n" +
		"  ```sql\n  ALTER TABLE accounts\n      ADD CONSTRAINT nonnegative_balance CHECK (balance >= 0);\n  ```\n"
	if !strings.Contains(body, want) {
		t.Fatalf("missing finding and its measured SQL:\n%s", body)
	}
	for _, unwanted := range []string{"statement 0", "unrelated statement"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("unexpected %q in comment:\n%s", unwanted, body)
		}
	}
}

func TestCommentOmitsSQLWhenTheFindingHasNoMatchingStatement(t *testing.T) {
	for _, tc := range []struct {
		name      string
		migration string
		index     int
		sql       string
	}{
		{name: "migration-wide finding", migration: "V2.sql", index: models.StatementScoped, sql: "SELECT 1;"},
		{name: "different migration", migration: "V1.sql", index: 0, sql: "SELECT 1;"},
		{name: "missing measurement", migration: "V2.sql", index: 9, sql: "SELECT 1;"},
		{name: "empty SQL", migration: "V2.sql", index: 0, sql: " \n\t"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			finding := models.RuntimeFinding{
				Type: "FINDING", Message: "Original diagnostic",
				OperationID: models.OperationID{MigrationID: tc.migration, StatementIndex: tc.index},
			}
			body := renderCommentFinding(finding, &models.RuntimeAnalysisResult{
				MigrationID: "V2.sql", Statements: []models.StatementMeasurement{{StatementIndex: 0, SQL: tc.sql}},
			})
			if body != "- ` FINDING `: Original diagnostic\n" {
				t.Fatalf("must retain the diagnostic without inventing SQL:\n%s", body)
			}
		})
	}
}

func TestCommentKeepsUnmodelledSQLLiteralInsideItsFence(t *testing.T) {
	const sql = "UPDATE accounts\nSET name = 'first line\n```\n</code><script>alert(1)</script> @someone\nlast line';"
	body := renderCommentFinding(models.RuntimeFinding{
		Type: models.FindingStatementFailed, Message: "statement 3 failed: invalid value",
		OperationID: models.OperationID{MigrationID: "V2.sql", StatementIndex: 3},
	}, &models.RuntimeAnalysisResult{
		MigrationID: "V2.sql", Statements: []models.StatementMeasurement{{StatementIndex: 3, SQL: sql}},
	})
	want := "\n  ````sql\n  " + strings.ReplaceAll(sql, "\n", "\n  ") + "\n  ````\n"
	if !strings.Contains(body, want) || !strings.Contains(body, "This statement failed: invalid value") {
		t.Fatalf("SQL must keep its literal contents inside an unbroken code fence:\n%s", body)
	}
}

func TestCommentBoundsSQLWithoutDroppingTheOtherFindings(t *testing.T) {
	result := &models.RuntimeAnalysisResult{
		MigrationID: "V2.sql", Verdict: models.RuntimeCompleted,
		Statements: []models.StatementMeasurement{{StatementIndex: 0, SQL: strings.Repeat("界", 2000) + "END_OF_SQL"}},
	}
	for i := range 11 {
		result.Findings = append(result.Findings, models.RuntimeFinding{
			Type: fmt.Sprintf("FINDING_%d", i), Message: "statement 0 needs attention",
			OperationID: models.OperationID{MigrationID: "V2.sql", StatementIndex: 0},
		})
	}
	body := RenderPRComment(Summary{Migrations: []Migration{{Name: "V2.sql", Runtime: result}}}, testCommentRun())
	if !utf8.ValidString(body) || len(body) > 55_000 || strings.Contains(body, "END_OF_SQL") {
		t.Fatalf("SQL must be bounded without corrupting Unicode: %d bytes", len(body))
	}
	for _, want := range []string{"FINDING_9", "SQL truncated; see the workflow artifacts", "More findings are listed in the workflow artifacts", "[View full analysis]"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in bounded report", want)
		}
	}
}
