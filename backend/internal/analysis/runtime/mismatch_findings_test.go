package runtime

import (
	"strings"
	"testing"

	"migration-timeline/backend/internal/models"
)

func mismatchResult(measurement models.StatementMeasurement) models.RuntimeAnalysisResult {
	measurement.Verdict = models.RuntimeCompleted
	return models.RuntimeAnalysisResult{
		Seeded:     []models.SeededTable{{Table: "accounts", Rows: 1_000_000}},
		Statements: []models.StatementMeasurement{measurement},
		Retry:      models.RetryNotApplicable,
	}
}

// The prediction calling a rewrite cheap is the case that reaches production: it
// passes the offline gate and then rewrites the table.
func TestFindingsReportAPredictionThatUnderstatedTheCost(t *testing.T) {
	migration := &models.Migration{ID: "V9__add_token"}
	result := mismatchResult(models.StatementMeasurement{
		StatementIndex:     0,
		SQL:                "ALTER TABLE accounts ADD COLUMN token INT DEFAULT random()::int",
		PredictedClass:     models.MetadataOnly,
		ObservedClass:      models.TableRewrite,
		RewrittenRelations: []string{"accounts"},
		TuplesRead:         2_000_000,
	})

	findings := Findings(migration, result)

	if len(findings) != 1 || findings[0].Type != models.FindingClassUnderstated {
		t.Fatalf("expected one CLASS_UNDERSTATED finding, got %+v", findings)
	}
	if !strings.Contains(findings[0].Message, "accounts") || !strings.Contains(findings[0].Message, "2000000") {
		t.Errorf("expected the finding to carry its evidence, got %q", findings[0].Message)
	}
}

func TestFindingsReportAPredictionThatOverstatedTheCost(t *testing.T) {
	migration := &models.Migration{ID: "V6__widen_email"}
	result := mismatchResult(models.StatementMeasurement{
		StatementIndex: 0,
		SQL:            "ALTER TABLE accounts ALTER COLUMN email TYPE TEXT",
		PredictedClass: models.TableRewrite,
		ObservedClass:  models.MetadataOnly,
	})

	findings := Findings(migration, result)

	if len(findings) != 1 || findings[0].Type != models.FindingClassOverstated {
		t.Fatalf("expected one CLASS_OVERSTATED finding, got %+v", findings)
	}
}

func TestFindingsStaySilentWhenThePredictionAgreesOrIsUnmeasured(t *testing.T) {
	agreed := mismatchResult(models.StatementMeasurement{
		SQL:            "ALTER TABLE accounts ADD COLUMN note TEXT",
		PredictedClass: models.MetadataOnly,
		ObservedClass:  models.MetadataOnly,
	})
	if findings := Findings(&models.Migration{ID: "V2__add_note"}, agreed); len(findings) != 0 {
		t.Errorf("expected no finding when the two agree, got %+v", findings)
	}

	unmeasured := mismatchResult(models.StatementMeasurement{
		SQL:            "ALTER TABLE accounts ADD COLUMN note TEXT",
		PredictedClass: models.MetadataOnly,
	})
	if findings := Findings(&models.Migration{ID: "V2__add_note"}, unmeasured); len(findings) != 0 {
		t.Errorf("expected no finding without an observation to compare, got %+v", findings)
	}
}

// The lock finding follows the measurement where there is one: a metadata-only
// ALTER holds ACCESS EXCLUSIVE for a time that does not grow with the table, even
// when static analysis predicted a rewrite.
func TestFindingsPreferTheObservationOverThePredictionForTheLockGate(t *testing.T) {
	result := mismatchResult(models.StatementMeasurement{
		SQL:            "ALTER TABLE accounts ALTER COLUMN email TYPE TEXT",
		PredictedClass: models.TableRewrite,
		ObservedClass:  models.MetadataOnly,
		StrongestLock:  "AccessExclusiveLock",
		Locks:          []models.LockObservation{{Mode: "AccessExclusiveLock", Relation: "accounts"}},
	})

	for _, finding := range Findings(&models.Migration{ID: "V6__widen_email"}, result) {
		if finding.Type == models.FindingExclusiveLock {
			t.Errorf("expected no lock finding for a measured metadata-only alter, got %q", finding.Message)
		}
	}
}
