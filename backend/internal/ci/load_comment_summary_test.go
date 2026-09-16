package ci

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"migration-lab/backend/internal/ci/prcomment"
	"migration-lab/backend/internal/models"
)

func TestCommentShowsConciseReviewSummary(t *testing.T) {
	run := testCommentRun()
	output := t.TempDir()
	target := "V2__add_account_status.sql"
	saveCommentFixture(t, output, "plan.json", migrationPlan{Head: run.Head, Directory: "db/migrations", Targets: []string{target}})
	saveCommentFixture(t, output, "results.json", []migrationResult{{Migration: target, Result: target + ".json"}})
	saveCommentFixture(t, output, target+".json", map[string]any{"runtimeResult": models.RuntimeAnalysisResult{
		MigrationID: target, Verdict: models.RuntimeCompleted, DeadlineMs: 5000,
		Statements: []models.StatementMeasurement{
			{DurationMs: 12, PredictedClass: models.TableRewrite, ObservedClass: models.MetadataOnly},
			{DurationMs: 3, PredictedClass: models.MetadataOnly, ObservedClass: models.MetadataOnly},
		},
		Seeded:  []models.SeededTable{{Table: "accounts", Rows: 1_000_000}},
		Message: "all 2 statements completed in 15 ms",
	}})
	body := prcomment.RenderPRComment(loadCommentSummary(output, run.Head), run)
	want := "## Migration runtime analysis\n\n" +
		"### [✅` " + target + " `](https://github.com/owner/repo/pull/7/files#diff-f71ba626fed0ac9e5c7e61d9722162e06e8cabcbc644d8250c35e21e06361207)\n\n" +
		"- No runtime concerns found.\n" +
		"- Performance class: ` METADATA_ONLY `\n" +
		"- [View full analysis](https://github.com/owner/repo/actions/runs/100/attempts/1)\n"
	if !strings.Contains(body, want) {
		t.Errorf("missing %q in comment:\n%s", want, body)
	}
	for _, unwanted := range []string{
		"Commit:", run.Head, "Directory:", "db/migrations", "Deadline:", "5000", "1000000", "**Seeding**",
		"all 2 statements", "Execution time is the sum", "findings alone do not fail", "**Verdict:**", "COMPLETED",
		"Recommendation:", "Execution:", "Predicted class:", "Observed class:", "TABLE_REWRITE",
	} {
		if strings.Contains(body, unwanted) {
			t.Errorf("unwanted detail %q in comment:\n%s", unwanted, body)
		}
	}
}

func TestCommentSummaryPreservesGitPathsWithoutRuntimeResults(t *testing.T) {
	for _, tc := range []struct{ directory, target, want string }{
		{".", "V2__root.sql", "V2__root.sql"},
		{"db/with space", "V2__`quoted``name [x] @someone 世.sql", "db/with space/V2__`quoted``name [x] @someone 世.sql"},
		{"", "V2__missing_directory.sql", ""},
		{"db", "../V2__invalid.sql", ""},
	} {
		t.Run(tc.target, func(t *testing.T) {
			output := t.TempDir()
			run := testCommentRun()
			saveCommentFixture(t, output, "plan.json", migrationPlan{Head: run.Head, Directory: tc.directory, Targets: []string{tc.target}})
			summary := loadCommentSummary(output, run.Head)
			if len(summary.Migrations) != 1 || summary.Migrations[0].Path != tc.want {
				t.Fatalf("wanted Git path %q, got %+v", tc.want, summary.Migrations)
			}
		})
	}
}

func TestCommentHandlesFailuresAndPartialArtifacts(t *testing.T) {
	for _, tc := range []struct {
		name   string
		setup  func(*testing.T, string, prcomment.Run)
		status string
		want   string
	}{
		{name: "early build failure", status: "failure", want: "selection plan is unavailable"},
		{name: "wrong commit", status: "success", want: "does not match", setup: func(t *testing.T, output string, _ prcomment.Run) {
			saveCommentFixture(t, output, "plan.json", migrationPlan{Head: "old"})
		}},
		{name: "missing results", status: "failure", want: "No runtime result was collected", setup: func(t *testing.T, output string, run prcomment.Run) {
			saveCommentFixture(t, output, "plan.json", migrationPlan{Head: run.Head, Targets: []string{"V2__add.sql"}})
		}},
		{name: "timeout with partial JSON", status: "failure", want: "CLI exit code 124", setup: func(t *testing.T, output string, run prcomment.Run) {
			saveCommentFixture(t, output, "plan.json", migrationPlan{Head: run.Head, Targets: []string{"V2__add.sql"}})
			saveCommentFixture(t, output, "results.json", []migrationResult{{Migration: "V2__add.sql", ExitCode: 124}})
			if err := os.WriteFile(filepath.Join(output, "V2__add.sql.json"), []byte(`{"runtimeResult":`), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "wrong migration in JSON", status: "success", want: "Runtime output is missing or invalid", setup: func(t *testing.T, output string, run prcomment.Run) {
			saveCommentFixture(t, output, "plan.json", migrationPlan{Head: run.Head, Targets: []string{"V2__add.sql"}})
			saveCommentFixture(t, output, "results.json", []migrationResult{{Migration: "V2__add.sql"}})
			saveCommentFixture(t, output, "V2__add.sql.json", map[string]any{"runtimeResult": models.RuntimeAnalysisResult{MigrationID: "other", Verdict: models.RuntimeCompleted}})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output := t.TempDir()
			run := testCommentRun()
			run.Status = tc.status
			if tc.setup != nil {
				tc.setup(t, output, run)
			}
			body := prcomment.RenderPRComment(loadCommentSummary(output, run.Head), run)
			if !strings.Contains(body, tc.want) || !strings.Contains(body, "## Migration runtime analysis\n") || strings.Contains(body, "✅") || strings.Contains(body, "No runtime concerns found") {
				t.Fatalf("misleading failure summary:\n%s", body)
			}
		})
	}
}

func TestCommentIncludesAllTargetsAfterRuntimeFailure(t *testing.T) {
	output := t.TempDir()
	run := testCommentRun()
	run.Status = "failure"
	targets := []string{"V2__first.sql", "V10__second.sql", "V11__unattempted.sql"}
	saveCommentFixture(t, output, "plan.json", migrationPlan{Head: run.Head, Targets: targets})
	saveCommentFixture(t, output, "results.json", []migrationResult{{Migration: targets[0], ExitCode: 1}, {Migration: targets[1]}})
	for i, verdict := range []models.RuntimeVerdict{models.RuntimeExceedsDeadline, models.RuntimeCompleted} {
		saveCommentFixture(t, output, targets[i]+".json", map[string]any{"runtimeResult": models.RuntimeAnalysisResult{MigrationID: targets[i], Verdict: verdict}})
	}
	body := prcomment.RenderPRComment(loadCommentSummary(output, run.Head), run)
	for _, want := range []string{"Analysis did not finish successfully", targets[0], targets[1], targets[2], "Do not merge", "exceeded the runtime limit", "Review before merging", "No runtime result was collected"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q:\n%s", want, body)
		}
	}
	for _, target := range targets {
		if !strings.Contains(body, "### ⚠️` "+target+" `\n") {
			t.Errorf("missing migration heading for %s:\n%s", target, body)
		}
	}
	if strings.Index(body, targets[0]) > strings.Index(body, targets[1]) {
		t.Fatal("comment should preserve the plan's numeric migration order")
	}
}

func TestCommentUpdatesWhenAddedMigrationIsRemoved(t *testing.T) {
	output := t.TempDir()
	run := testCommentRun()
	saveCommentFixture(t, output, "plan.json", migrationPlan{Head: run.Head, Targets: []string{}, ExistingChanges: []migrationChange{{}}})
	body := prcomment.RenderPRComment(loadCommentSummary(output, run.Head), run)
	for _, want := range []string{"<!-- migration-lab:runtime-analysis -->", "## Migration runtime analysis\n", "- No new migrations", "runtime analysis was skipped", "- 1 existing migration change(s)", "- [View full analysis]"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q:\n%s", want, body)
		}
	}
}

func testCommentRun() prcomment.Run {
	return prcomment.Run{Repository: "owner/repo", Number: 7, Head: strings.Repeat("a", 40), RunID: 100, Attempt: 1, Status: "success"}
}

func saveCommentFixture(t *testing.T, output, name string, value any) {
	t.Helper()
	if err := writeJSON(filepath.Join(output, name), value); err != nil {
		t.Fatal(err)
	}
}
