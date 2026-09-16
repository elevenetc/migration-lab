package prcomment

import (
	"fmt"
	"html"
	"slices"
	"strings"

	"migration-lab/backend/internal/models"
)

const commentMarker = "<!-- migration-lab:runtime-analysis -->"

// Run identifies the pull request and workflow attempt associated with a comment.
type Run struct {
	Repository string
	Number     int
	Head       string
	RunID      int64
	Attempt    int64
	Status     string
}

// RenderPRComment formats a bounded Markdown summary with publisher metadata.
func RenderPRComment(summary Summary, run Run) string {
	body := commentMarker + fmt.Sprintf("\n<!-- migration-lab-run: %d %d -->\n", run.RunID, run.Attempt)
	if len(summary.Migrations) == 0 {
		emoji := "ℹ️"
		if run.Status != "success" || summary.Problem != "" {
			emoji = "⚠️"
		}
		body += "## " + emoji + " Migration runtime analysis\n\n"
	}
	if run.Status != "success" {
		body += "- **Analysis did not finish successfully.** See the workflow for diagnostics.\n"
	}
	if summary.Problem != "" {
		body += "- **Analysis incomplete — rerun before merging.** " + commentText(summary.Problem) + "\n"
	}
	if len(summary.Migrations) == 0 && summary.Problem == "" {
		body += "- No new migrations; runtime analysis was skipped.\n"
	}
	if summary.ExistingChanges > 0 {
		body += fmt.Sprintf("- %d existing migration change(s) were recorded but not measured.\n", summary.ExistingChanges)
	}
	if run.Status != "success" || summary.Problem != "" || summary.ExistingChanges > 0 {
		body += "\n"
	}
	for _, migration := range summary.Migrations {
		section := renderCommentMigration(migration)
		// GitHub limits comment bodies to 65,536 characters. Bound bytes as well
		// and keep whole sections so truncation never breaks Markdown formatting.
		if len(body)+len(section) > 55_000 {
			body += "- Additional migration results are available in the workflow artifacts.\n"
			break
		}
		body += section + "\n"
	}
	body = strings.TrimRight(body, "\n") + "\n"
	body += fmt.Sprintf("- [View full analysis](https://github.com/%s/actions/runs/%d/attempts/%d)\n", run.Repository, run.RunID, run.Attempt)
	return body
}

func renderCommentMigration(migration Migration) string {
	recommendation := commentRecommendation(migration)
	classes := commentClasses{}
	if migration.Problem == "" && migration.Runtime != nil {
		classes = commentPerformance(migration.Runtime.Statements)
	}
	emoji := "⚠️"
	switch {
	case classes.Observed == models.TableRewrite:
		emoji = "🚨"
	case classes.Observed == models.MetadataOnly && recommendation.NoConcerns:
		emoji = "✅"
	}
	body := "## " + emoji + " Migration runtime analysis - " + commentCode(migration.Name) + "\n\n"
	body += "- " + commentText(recommendation.Message) + "\n"
	if migration.Problem != "" || migration.Runtime == nil {
		return body + "- **Result unavailable:** " + commentText(migration.Problem) + "\n"
	}
	runtime := migration.Runtime
	body += "- Performance class: " + renderCommentClass(classes.Observed, classes.ObservedComplete) + "\n"
	if runtime.Verdict != models.RuntimeCompleted && runtime.Message != "" {
		body += "- " + commentText(runtime.Message) + "\n"
	}
	for i, finding := range runtime.Findings {
		if i == 10 {
			body += "- More findings are listed in the workflow artifacts.\n"
			break
		}
		body += renderCommentFinding(finding, runtime)
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
			body += "- More seeding failures are listed in the workflow artifacts.\n"
			break
		}
		body += "- **Seeding failed:** " + commentCode(table.Table) + " — " + commentText(table.Error) + "\n"
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
