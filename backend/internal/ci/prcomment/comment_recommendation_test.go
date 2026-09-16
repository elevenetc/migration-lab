package prcomment

import (
	"strings"
	"testing"

	"migration-lab/backend/internal/models"
)

func TestCommentRecommendationUsesRiskAndEvidence(t *testing.T) {
	for _, tc := range []struct {
		name      string
		change    func(*Migration)
		want      string
		wantEmoji string
	}{
		{name: "clean metadata operation", want: "No runtime concerns found", wantEmoji: "✅"},
		{name: "metadata-only exclusive lock", want: "No runtime concerns found", wantEmoji: "✅", change: func(m *Migration) {
			m.Runtime.Statements[0].StrongestLock = "AccessExclusiveLock"
			m.Runtime.Statements[0].Locks = []models.LockObservation{{Mode: "AccessExclusiveLock", Relation: "accounts"}}
		}},
		{name: "reader-blocking lock", want: "Review before merging — reader-blocking lock duration", change: func(m *Migration) {
			m.Runtime.Findings = []models.RuntimeFinding{{Type: models.FindingExclusiveLock}}
		}},
		{name: "fast scan with no findings", want: "Review before merging — the migration scans table data", change: func(m *Migration) {
			m.Runtime.Statements[0].ObservedClass = models.DataScanning
			m.Runtime.Statements[0].PredictedClass = models.DataScanning
		}},
		{name: "fast rewrite with no findings", want: "Review before merging — the migration rewrites table data", wantEmoji: "🚨", change: func(m *Migration) {
			m.Runtime.Statements[0].ObservedClass = models.TableRewrite
		}},
		{name: "mixed observed classes", want: "Review before merging — the migration rewrites table data", wantEmoji: "🚨", change: func(m *Migration) {
			m.Runtime.Statements = append(m.Runtime.Statements,
				models.StatementMeasurement{ObservedClass: models.TableRewrite},
				models.StatementMeasurement{ObservedClass: models.DataScanning},
			)
		}},
		{name: "rewrite with reader-blocking finding", want: "Review before merging — reader-blocking lock duration", wantEmoji: "🚨", change: func(m *Migration) {
			m.Runtime.Statements[0].ObservedClass = models.TableRewrite
			m.Runtime.Findings = []models.RuntimeFinding{{Type: models.FindingExclusiveLock}}
		}},
		{name: "partial rewrite observations", want: "Review before merging — performance could not be verified", wantEmoji: "🚨", change: func(m *Migration) {
			m.Runtime.Statements[0].ObservedClass = models.TableRewrite
			m.Runtime.Statements = append(m.Runtime.Statements, models.StatementMeasurement{})
		}},
		{name: "pessimistic prediction alone is informative", want: "No runtime concerns found", wantEmoji: "✅", change: func(m *Migration) {
			m.Runtime.Statements[0].PredictedClass = models.TableRewrite
			m.Runtime.Findings = []models.RuntimeFinding{{Type: models.FindingClassOverstated}}
		}},
		{name: "unknown findings need review", want: "Review before merging", change: func(m *Migration) {
			m.Runtime.Findings = []models.RuntimeFinding{{Type: "NEW_RISK"}}
		}},
		{name: "failed execution", want: "Do not merge — the migration failed", change: func(m *Migration) {
			m.Runtime.Verdict = models.RuntimeFailed
			m.ExitCode = 1
		}},
		{name: "deadline", want: "Do not merge — the migration exceeded", change: func(m *Migration) {
			m.Runtime.Verdict = models.RuntimeExceedsDeadline
		}},
		{name: "manual cleanup", want: "Do not merge — the migration cannot be retried", change: func(m *Migration) {
			m.Runtime.Retry = models.RetryManualCleanup
		}},
		{name: "retry failure loop", want: "Do not merge — the migration cannot be retried", change: func(m *Migration) {
			m.Runtime.Retry = models.RetryFailureLoop
		}},
		{name: "failure finding takes priority over blocking", want: "Do not merge", change: func(m *Migration) {
			m.Runtime.Findings = []models.RuntimeFinding{{Type: models.FindingExclusiveLock}, {Type: models.FindingStatementFailed}}
		}},
		{name: "failed seeding", want: "Analysis incomplete — resolve seeding coverage", change: func(m *Migration) {
			m.Runtime.Seeded[0].Error = "failed"
		}},
		{name: "empty seeded table", want: "Analysis incomplete — resolve seeding coverage", change: func(m *Migration) {
			m.Runtime.Seeded[0].Rows = 0
		}},
		{name: "seed failure finding without table entry", want: "Analysis incomplete — resolve seeding coverage", change: func(m *Migration) {
			m.Runtime.Findings = []models.RuntimeFinding{{Type: models.FindingSeedFailed}}
		}},
		{name: "missing runtime result", want: "Analysis incomplete", change: func(m *Migration) {
			m.Runtime = nil
		}},
		{name: "invalid runtime result", want: "Analysis incomplete", change: func(m *Migration) {
			m.Problem = "Runtime output is invalid."
		}},
		{name: "failed CLI despite completed execution", want: "Analysis incomplete", change: func(m *Migration) {
			m.ExitCode = 1
		}},
		{name: "unknown verdict", want: "Analysis incomplete", change: func(m *Migration) {
			m.Runtime.Verdict = "UNKNOWN"
		}},
		{name: "missing observed class", want: "Review before merging — performance could not be verified", change: func(m *Migration) {
			m.Runtime.Statements[0].ObservedClass = ""
		}},
		{name: "partial observations", want: "Review before merging — performance could not be verified", change: func(m *Migration) {
			m.Runtime.Statements = append(m.Runtime.Statements, models.StatementMeasurement{})
		}},
		{name: "no measured statements", want: "Review before merging — performance could not be verified", change: func(m *Migration) {
			m.Runtime.Statements = nil
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			migration := Migration{Runtime: &models.RuntimeAnalysisResult{
				Verdict: models.RuntimeCompleted,
				Seeded:  []models.SeededTable{{Table: "accounts", Rows: 1_000_000}},
				Statements: []models.StatementMeasurement{{
					DurationMs: 3, PredictedClass: models.MetadataOnly, ObservedClass: models.MetadataOnly,
				}},
			}}
			if tc.change != nil {
				tc.change(&migration)
			}
			if advice := commentRecommendation(migration); !strings.HasPrefix(advice.Message, tc.want) {
				t.Fatalf("wanted %q, got %q", tc.want, advice.Message)
			}
			emoji := tc.wantEmoji
			if emoji == "" {
				emoji = "⚠️"
			}
			if body := renderCommentMigration(migration, testCommentRun()); !strings.HasPrefix(body, "### "+emoji+"`") {
				t.Fatalf("heading must reflect observed performance and runtime concerns:\n%s", body)
			}
		})
	}
}

func TestCommentPerformanceReportsWorstClassesAndMissingCoverage(t *testing.T) {
	statements := []models.StatementMeasurement{
		{PredictedClass: models.MetadataOnly, ObservedClass: models.DataScanning},
		{PredictedClass: models.TableRewrite, ObservedClass: models.MetadataOnly},
	}
	classes := commentPerformance(statements)
	if classes.Observed != models.DataScanning || !classes.ObservedComplete {
		t.Fatalf("must show the worst observed class: %+v", classes)
	}
	classes = commentPerformance(append(statements, models.StatementMeasurement{}))
	if classes.Observed != models.DataScanning || classes.ObservedComplete {
		t.Fatalf("unknown statements must not be classified as metadata-only: %+v", classes)
	}
	if text := renderCommentClass(classes.Observed, classes.ObservedComplete); !strings.Contains(text, "DATA_SCANNING") || !strings.Contains(text, "(partial)") {
		t.Fatalf("partial coverage should be visible: %s", text)
	}
	classes = commentPerformance([]models.StatementMeasurement{{ObservedClass: "NEW_CLASS"}})
	if text := renderCommentClass(classes.Observed, classes.ObservedComplete); text != "Unavailable" {
		t.Fatalf("unknown classifications must remain unavailable: %s", text)
	}
}

func TestCommentPreservesMigrationFailureMessage(t *testing.T) {
	body := renderCommentMigration(Migration{Name: "V2__fail.sql", ExitCode: 1, Runtime: &models.RuntimeAnalysisResult{
		Verdict: models.RuntimeFailed, Message: "statement 0 failed: duplicate column",
	}}, testCommentRun())
	if !strings.Contains(body, "Do not merge") || !strings.Contains(body, "duplicate column") || strings.Contains(body, "Deadline:") {
		t.Fatalf("failure advice must retain its explanation:\n%s", body)
	}
}
