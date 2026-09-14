package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"migration-lab/backend/internal/models"
)

func TestCommentShowsConciseReviewSummary(t *testing.T) {
	run := testCommentRun()
	output := t.TempDir()
	target := "V2__add_account_status.sql"
	saveCommentFixture(t, output, "plan.json", migrationPlan{Head: run.Head, Targets: []string{target}})
	saveCommentFixture(t, output, "results.json", []migrationResult{{Migration: target, Result: target + ".json"}})
	saveCommentFixture(t, output, target+".json", map[string]any{"runtimeResult": models.RuntimeAnalysisResult{
		MigrationID: target, Verdict: models.RuntimeCompleted, DeadlineMs: 5000,
		Statements: []models.StatementMeasurement{
			{DurationMs: 12, PredictedClass: models.MetadataOnly, ObservedClass: models.MetadataOnly},
			{DurationMs: 3, PredictedClass: models.MetadataOnly, ObservedClass: models.MetadataOnly},
		},
		Seeded:  []models.SeededTable{{Table: "accounts", Rows: 1_000_000}},
		Message: "all 2 statements completed in 15 ms",
	}})
	body := renderPRComment(loadCommentSummary(output, run.Head), run)
	for _, want := range []string{
		"## Migration runtime analysis\n", target, "**Recommendation:** No runtime concerns found.",
		"**Execution:** 15 ms", "**Predicted class:** ` METADATA_ONLY `", "**Observed class:** ` METADATA_ONLY `",
		"[View full analysis]", "/actions/runs/100/attempts/1",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in comment:\n%s", want, body)
		}
	}
	for _, unwanted := range []string{
		"Commit:", run.Head, "Directory:", "db/migrations", "Deadline:", "5000", "1000000", "**Seeding**",
		"all 2 statements", "Execution time is the sum", "findings alone do not fail", "**Verdict:**", "COMPLETED",
	} {
		if strings.Contains(body, unwanted) {
			t.Errorf("unwanted detail %q in comment:\n%s", unwanted, body)
		}
	}
}

func TestCommentShowsFindingsAndFailedSeedingEvenWhenCompleted(t *testing.T) {
	body := renderPRComment(commentSummary{Migrations: []commentMigration{{
		Name: "V2__change.sql", Runtime: &models.RuntimeAnalysisResult{
			Verdict: models.RuntimeCompleted,
			Seeded:  []models.SeededTable{{Table: "accounts", Error: "could not seed accounts"}},
			Findings: []models.RuntimeFinding{
				{Type: models.FindingSeedFailed, Message: "Measurements used an unseeded table"},
				{Type: models.FindingExclusiveLock, Message: "Reader-blocking lock duration grows with table size"},
			},
		},
	}}}, testCommentRun())
	for _, want := range []string{"Migration runtime analysis", "Analysis incomplete", "**Seeding failed:**", "could not seed accounts", "SEED_FAILED", "EXCLUSIVE_LOCK_HELD", "Reader-blocking lock duration grows with table size"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in comment:\n%s", want, body)
		}
	}
	if strings.Contains(body, "1000000") || strings.Contains(body, "No runtime concerns found") {
		t.Fatalf("comment must not invent seeding coverage or hide findings:\n%s", body)
	}
}

func TestCommentHandlesFailuresAndPartialArtifacts(t *testing.T) {
	for _, tc := range []struct {
		name   string
		setup  func(*testing.T, string, commentRun)
		status string
		want   string
	}{
		{name: "early build failure", status: "failure", want: "selection plan is unavailable"},
		{name: "wrong commit", status: "success", want: "does not match", setup: func(t *testing.T, output string, _ commentRun) {
			saveCommentFixture(t, output, "plan.json", migrationPlan{Head: "old"})
		}},
		{name: "missing results", status: "failure", want: "No runtime result was collected", setup: func(t *testing.T, output string, run commentRun) {
			saveCommentFixture(t, output, "plan.json", migrationPlan{Head: run.Head, Targets: []string{"V2__add.sql"}})
		}},
		{name: "timeout with partial JSON", status: "failure", want: "CLI exit code 124", setup: func(t *testing.T, output string, run commentRun) {
			saveCommentFixture(t, output, "plan.json", migrationPlan{Head: run.Head, Targets: []string{"V2__add.sql"}})
			saveCommentFixture(t, output, "results.json", []migrationResult{{Migration: "V2__add.sql", ExitCode: 124}})
			if err := os.WriteFile(filepath.Join(output, "V2__add.sql.json"), []byte(`{"runtimeResult":`), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "wrong migration in JSON", status: "success", want: "Runtime output is missing or invalid", setup: func(t *testing.T, output string, run commentRun) {
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
			body := renderPRComment(loadCommentSummary(output, run.Head), run)
			if !strings.Contains(body, tc.want) || strings.Contains(body, "No runtime concerns found") {
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
	body := renderPRComment(loadCommentSummary(output, run.Head), run)
	for _, want := range []string{"Analysis did not finish successfully", targets[0], targets[1], targets[2], "Do not merge", "exceeded the runtime limit", "Review before merging", "No runtime result was collected"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q:\n%s", want, body)
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
	body := renderPRComment(loadCommentSummary(output, run.Head), run)
	for _, want := range []string{commentMarker, "No new migrations", "runtime analysis was skipped", "1 existing migration change(s)"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q:\n%s", want, body)
		}
	}
}

func TestCommentEscapesRepositoryTextAndBoundsLargeReports(t *testing.T) {
	untrusted := "</code><script>alert(1)</script> @someone\n[link](https://example.com) `tick` *bold*"
	summary := commentSummary{Migrations: []commentMigration{{Name: untrusted, Runtime: &models.RuntimeAnalysisResult{
		Verdict: models.RuntimeCompleted, Findings: []models.RuntimeFinding{{Type: "TYPE", Message: untrusted}},
	}}}}
	body := renderPRComment(summary, testCommentRun())
	for _, unwanted := range []string{"<script>", "@someone", "[link]", "`tick`", "*bold*"} {
		// The migration filename is in a code span; Markdown syntax there is inert.
		findingText := strings.Split(body, "**Runtime findings")[1]
		if strings.Contains(findingText, unwanted) {
			t.Errorf("unescaped text %q:\n%s", unwanted, body)
		}
	}
	for range 1000 {
		summary.Migrations = append(summary.Migrations, commentMigration{Name: strings.Repeat("世", 1000), Problem: strings.Repeat("<>&", 1000)})
	}
	body = renderPRComment(summary, testCommentRun())
	if len(body) > 60_000 || !utf8.ValidString(body) || !strings.Contains(body, "Additional migration results") {
		t.Fatalf("large comment is not safely truncated: %d bytes", len(body))
	}
}

func TestCommentEscapingPreservesEntitiesAndCodeBoundaries(t *testing.T) {
	text := commentText(`column "status" can't contain <markup> or #mentions`)
	if strings.Contains(text, `&\#`) || !strings.Contains(text, "&#34;status&#34;") || !strings.Contains(text, `\#mentions`) {
		t.Fatalf("HTML entities must survive Markdown escaping: %s", text)
	}
	if code := commentCode("V2__`quoted``name.sql"); code != "``` V2__`quoted``name.sql ```" {
		t.Fatalf("backticks in a filename must not end its code span: %s", code)
	}
}

func testCommentRun() commentRun {
	return commentRun{Repository: "owner/repo", Number: 7, Head: strings.Repeat("a", 40), RunID: 100, Attempt: 1, Status: "success"}
}

func saveCommentFixture(t *testing.T, output, name string, value any) {
	t.Helper()
	if err := writeJSON(filepath.Join(output, name), value); err != nil {
		t.Fatal(err)
	}
}
