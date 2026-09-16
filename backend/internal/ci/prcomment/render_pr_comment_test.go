package prcomment

import (
	"strings"
	"testing"
	"unicode/utf8"

	"migration-lab/backend/internal/models"
)

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
	for _, want := range []string{"## ⚠️ Migration runtime analysis - ` V2__change.sql `", "- Analysis incomplete", "- Performance class: Unavailable", "- **Seeding failed:**", "could not seed accounts", "SEED_FAILED", "EXCLUSIVE_LOCK_HELD", "Reader-blocking lock duration grows with table size"} {
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
