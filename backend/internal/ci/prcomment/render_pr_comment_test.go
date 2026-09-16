package prcomment

import (
	"strings"
	"testing"
	"unicode/utf8"

	"migration-lab/backend/internal/models"
)

func TestCommentHasOneTitleAndLinkedMigrationSubtitles(t *testing.T) {
	migrations := []Migration{
		{Name: "V2__add_account_status.sql", Path: "db/migrations/V2__add_account_status.sql"},
		{Name: "V2__root.sql", Path: "V2__root.sql"},
		{Name: "V2__`quoted``name [x] @someone 世.sql", Path: "db/with space/V2__`quoted``name [x] @someone 世.sql"},
	}
	classes := []models.PerformanceClass{models.MetadataOnly, models.DataScanning, models.TableRewrite}
	for i := range migrations {
		migrations[i].Runtime = &models.RuntimeAnalysisResult{
			Verdict:    models.RuntimeCompleted,
			Statements: []models.StatementMeasurement{{ObservedClass: classes[i]}},
		}
	}
	body := RenderPRComment(Summary{Migrations: migrations}, testCommentRun())
	if strings.Count(body, "Migration runtime analysis") != 1 || !strings.Contains(body, "\n## Migration runtime analysis\n\n") {
		t.Fatalf("expected a single main title:\n%s", body)
	}
	wantHeadings := []string{
		"### [✅` V2__add_account_status.sql `](https://github.com/owner/repo/pull/7/files#diff-f71ba626fed0ac9e5c7e61d9722162e06e8cabcbc644d8250c35e21e06361207)",
		"### [⚠️` V2__root.sql `](https://github.com/owner/repo/pull/7/files#diff-102bb1a5c568b5924b1626dd7da90bc98b1e6cb7afe643cf4181145c5418d599)",
		"### [🚨``` V2__`quoted``name [x] @\u200bsomeone 世.sql ```](https://github.com/owner/repo/pull/7/files#diff-40261c023d2a62e19ae10a3224d89537740dd66e42957aed923fc6d92fc8c4d4)",
	}
	previous := -1
	for _, want := range wantHeadings {
		index := strings.Index(body, want+"\n\n")
		if index <= previous {
			t.Fatalf("missing or out-of-order subtitle %q:\n%s", want, body)
		}
		previous = index
	}
}

func TestCommentShowsFindingsAndFailedSeedingEvenWhenCompleted(t *testing.T) {
	body := RenderPRComment(Summary{Migrations: []Migration{{
		Name: "V2__change.sql", Runtime: &models.RuntimeAnalysisResult{
			Verdict: models.RuntimeCompleted,
			Seeded:  []models.SeededTable{{Table: "accounts", Error: "could not seed accounts"}},
			Findings: []models.RuntimeFinding{
				{Type: models.FindingSeedFailed, Message: "Measurements used an unseeded table"},
				{Type: models.FindingExclusiveLock, Message: "Reader-blocking lock duration grows with table size"},
			},
		},
	}}}, testCommentRun())
	for _, want := range []string{"## Migration runtime analysis\n", "### ⚠️` V2__change.sql `", "- Analysis incomplete", "- Performance class: Unavailable", "- **Seeding failed:**", "could not seed accounts", "SEED_FAILED", "EXCLUSIVE_LOCK_HELD", "Reader-blocking lock duration grows with table size"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in comment:\n%s", want, body)
		}
	}
	if strings.Contains(body, "1000000") || strings.Contains(body, "No runtime concerns found") {
		t.Fatalf("comment must not invent seeding coverage or hide findings:\n%s", body)
	}
}

func TestCommentEscapesRepositoryTextAndBoundsLargeReports(t *testing.T) {
	untrusted := "</code><script>alert(1)</script> @someone\n[link](https://example.com) `tick` *bold*"
	summary := Summary{Migrations: []Migration{{Name: untrusted, Runtime: &models.RuntimeAnalysisResult{
		Verdict: models.RuntimeCompleted, Findings: []models.RuntimeFinding{{Type: "TYPE", Message: untrusted}},
	}}}}
	body := RenderPRComment(summary, testCommentRun())
	for _, unwanted := range []string{"<script>", "@someone", "[link]", "`tick`", "*bold*"} {
		// The migration filename is in a code span; Markdown syntax there is inert.
		findingText := strings.Split(body, "- ` TYPE `:")[1]
		if strings.Contains(findingText, unwanted) {
			t.Errorf("unescaped text %q:\n%s", unwanted, body)
		}
	}
	for range 1000 {
		summary.Migrations = append(summary.Migrations, Migration{Name: strings.Repeat("世", 1000), Problem: strings.Repeat("<>&", 1000)})
	}
	body = RenderPRComment(summary, testCommentRun())
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

func testCommentRun() Run {
	return Run{Repository: "owner/repo", Number: 7, Head: strings.Repeat("a", 40), RunID: 100, Attempt: 1, Status: "success"}
}
