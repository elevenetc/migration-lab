package main

import (
	"fmt"
	"html"
	"slices"
	"strings"

	"migration-lab/backend/internal/models"
)

const commentMarker = "<!-- migration-lab:runtime-analysis -->"

type commentRun struct {
	Repository string
	Number     int
	Head       string
	RunID      int64
	Attempt    int64
	Status     string
}

func renderPRComment(summary commentSummary, run commentRun) string {
	body := commentMarker + fmt.Sprintf("\n<!-- migration-lab-run: %d %d -->\n", run.RunID, run.Attempt)
	body += "## Migration runtime analysis\n\n"
	if run.Status != "success" {
		body += "**Analysis did not finish successfully.** See the workflow for diagnostics.\n\n"
	}
	if summary.Problem != "" {
		body += "**Analysis incomplete — rerun before merging.** " + commentText(summary.Problem) + "\n\n"
	}
	if len(summary.Migrations) == 0 && summary.Problem == "" {
		body += "No new migrations; runtime analysis was skipped.\n\n"
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
	body += fmt.Sprintf("[View full analysis](https://github.com/%s/actions/runs/%d/attempts/%d)\n", run.Repository, run.RunID, run.Attempt)
	return body
}

func renderCommentMigration(migration commentMigration) string {
	body := "### " + commentCode(migration.Name) + "\n\n"
	body += "**Recommendation:** " + commentText(commentRecommendation(migration)) + "\n\n"
	if migration.Problem != "" || migration.Runtime == nil {
		return body + "**Result unavailable:** " + commentText(migration.Problem) + "\n\n"
	}
	runtime := migration.Runtime
	var duration int64
	for _, statement := range runtime.Statements {
		duration += statement.DurationMs
	}
	classes := commentPerformance(runtime.Statements)
	body += "**Predicted class:** " + renderCommentClass(classes.Predicted, classes.PredictedComplete) +
		" · **Observed class:** " + renderCommentClass(classes.Observed, classes.ObservedComplete) + "\n\n"
	body += fmt.Sprintf("**Execution:** %d ms\n\n", duration)
	if runtime.Verdict != models.RuntimeCompleted && runtime.Message != "" {
		body += commentText(runtime.Message) + "\n\n"
	}
	if len(runtime.Findings) > 0 {
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
	// Seeding failures are normally findings. Retain diagnostics if an incomplete
	// result contains a failed seed without its corresponding finding.
	shown := 0
	for _, table := range runtime.Seeded {
		if table.Error == "" || slices.ContainsFunc(runtime.Findings, func(finding models.RuntimeFinding) bool {
			return finding.Type == models.FindingSeedFailed && finding.TableName == table.Table
		}) {
			continue
		}
		if shown == 10 {
			body += "More seeding failures are listed in the workflow artifacts.\n\n"
			break
		}
		body += "**Seeding failed:** " + commentCode(table.Table) + " — " + commentText(table.Error) + "\n\n"
		shown++
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
