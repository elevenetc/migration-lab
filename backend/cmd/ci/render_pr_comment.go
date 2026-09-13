package main

import (
	"fmt"
	"html"
	"strings"

	"migration-lab/backend/internal/models"
)

const commentMarker = "<!-- migration-lab:runtime-analysis -->"

type commentRun struct {
	Repository string
	Number     int
	Head       string
	Directory  string
	RunID      int64
	Attempt    int64
	Status     string
}

func renderPRComment(summary commentSummary, run commentRun) string {
	status := "Completed"
	if len(summary.Migrations) == 0 {
		status = "No new migrations"
	}
	if summary.Problem != "" {
		status = "Analysis incomplete"
	}
	for _, migration := range summary.Migrations {
		if migration.Problem != "" || migration.Runtime == nil || migration.ExitCode != 0 || migration.Runtime.Verdict != models.RuntimeCompleted {
			status = "Analysis incomplete"
		}
	}
	if run.Status != "success" {
		status = "Analysis " + run.Status
		if run.Status == "failure" {
			status = "Analysis failed"
		}
	}
	// Formatting into strings avoids unchecked writer errors in summary generation.
	body := commentMarker + fmt.Sprintf("\n<!-- migration-lab-run: %d %d -->\n", run.RunID, run.Attempt)
	body += "## Migration Lab — " + commentText(status) + "\n\n"
	body += "Commit: " + commentCode(run.Head) + " · Directory: " + commentCode(run.Directory) + "\n\n"
	body += fmt.Sprintf("[View workflow results and diagnostic artifacts](https://github.com/%s/actions/runs/%d/attempts/%d)\n\n", run.Repository, run.RunID, run.Attempt)
	if summary.Problem != "" {
		body += commentText(summary.Problem) + "\n\n"
	}
	if len(summary.Migrations) == 0 && summary.Problem == "" {
		body += "No added SQL migrations were selected; runtime analysis was skipped.\n\n"
	}
	if summary.ExistingChanges > 0 {
		body += fmt.Sprintf("%d existing migration change(s) were recorded but not measured.\n\n", summary.ExistingChanges)
	}
	for _, migration := range summary.Migrations {
		section := renderCommentMigration(migration)
		// GitHub limits comment bodies to 65,536 characters. Bound bytes as well
		// and keep whole sections so truncation never breaks Markdown formatting.
		if len(body)+len(section) > 55_000 {
			body += "Additional migration results are available in the workflow artifacts.\n\n"
			break
		}
		body += section
	}
	body += "Execution time is the sum of measured statements, excluding setup and seeding. Runtime findings alone do not fail the check.\n"
	return body
}

func renderCommentMigration(migration commentMigration) string {
	body := "### " + commentCode(migration.Name) + "\n\n"
	if migration.Problem != "" || migration.Runtime == nil {
		return body + "**Result unavailable:** " + commentText(migration.Problem) + "\n\n"
	}
	runtime := migration.Runtime
	var duration int64
	for _, statement := range runtime.Statements {
		duration += statement.DurationMs
	}
	body += fmt.Sprintf("**Verdict:** %s · **Execution:** %d ms · **Deadline:** %d ms\n\n", commentCode(string(runtime.Verdict)), duration, runtime.DeadlineMs)
	if migration.ExitCode != 0 {
		body += fmt.Sprintf("CLI exit code: %d.\n\n", migration.ExitCode)
	}
	if runtime.Message != "" {
		body += commentText(runtime.Message) + "\n\n"
	}
	if len(runtime.Seeded) == 0 {
		body += "**Seeding:** no existing tables were seeded.\n\n"
	} else {
		body += "**Seeding**\n\n"
		for i, table := range runtime.Seeded {
			if i == 10 {
				body += "- More tables are listed in the workflow artifacts.\n"
				break
			}
			if table.Error != "" {
				body += "- " + commentCode(table.Table) + ": **failed** — " + commentText(table.Error) + "\n"
			} else {
				body += fmt.Sprintf("- %s: %d rows\n", commentCode(table.Table), table.Rows)
			}
		}
		body += "\n"
	}
	if len(runtime.Findings) == 0 {
		body += "**Runtime findings:** none.\n\n"
	} else {
		body += fmt.Sprintf("**Runtime findings (%d)**\n\n", len(runtime.Findings))
		for i, finding := range runtime.Findings {
			if i == 10 {
				body += "- More findings are listed in the workflow artifacts.\n"
				break
			}
			body += "- " + commentCode(finding.Type) + ": " + commentText(finding.Message) + "\n"
		}
		body += "\n"
	}
	return body
}

func commentCode(value string) string {
	value = commentPlainText(value)
	longest, current := 0, 0
	for _, char := range value {
		if char == '`' {
			current++
			longest = max(longest, current)
		} else {
			current = 0
		}
	}
	fence := strings.Repeat("`", longest+1)
	return fence + " " + value + " " + fence
}

func commentText(value string) string {
	escaped := strings.NewReplacer("\\", "\\\\", "`", "\\`", "*", "\\*", "_", "\\_", "[", "\\[", "]", "\\]", "~", "\\~", "|", "\\|", "#", "\\#").Replace(commentPlainText(value))
	return html.EscapeString(escaped)
}

func commentPlainText(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	runes := []rune(value)
	if len(runes) > 600 {
		value = string(runes[:600]) + "…"
	}
	// Break @mentions in repository-controlled identifiers and database messages.
	// Callers either escape markup or enclose the result in a Markdown code span.
	return strings.ReplaceAll(value, "@", "@\u200b")
}
