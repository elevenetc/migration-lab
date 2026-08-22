package runtime

import (
	"context"
	"testing"
	"time"

	"migration-timeline/backend/internal/models"
)

// partitionedTimeline is the case the runtime pass exists for: a rewrite of a
// column on a partitioned parent, which takes ACCESS EXCLUSIVE on the parent and
// every partition under it.
func partitionedTimeline() []models.MigrationInfo {
	return []models.MigrationInfo{
		{ID: "V1__create_events", Timestamp: 1, SQL: `CREATE TABLE events (
			id SERIAL,
			year INT NOT NULL,
			note TEXT
		) PARTITION BY RANGE (year);`},
		{ID: "V2__create_events_2026", Timestamp: 2, SQL: `CREATE TABLE events_2026 PARTITION OF events
			FOR VALUES FROM (2026) TO (2027);`},
		{ID: "V3__widen_note", Timestamp: 3, SQL: `ALTER TABLE events ALTER COLUMN note TYPE VARCHAR(200);`},
	}
}

func TestAnalyseSeedsThePartitionAndSeesTheExclusiveLock(t *testing.T) {
	result, err := Analyse(context.Background(), Request{
		Migrations: partitionedTimeline(),
		Target:     "V3__widen_note",
		Rows:       20_000,
		Deadline:   2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("Analyse failed: %v", err)
	}

	if result.Verdict != models.RuntimeCompleted {
		t.Fatalf("expected the migration to complete inside two minutes, got %s (%s)", result.Verdict, result.Message)
	}
	if result.Version != "3" {
		t.Errorf("expected version 3, got %q", result.Version)
	}

	if len(result.Seeded) != 1 || result.Seeded[0].Table != "events_2026" {
		t.Fatalf("expected the partition to be seeded rather than the parent, got %+v", result.Seeded)
	}
	if result.Seeded[0].Error != "" {
		t.Fatalf("expected events_2026 to seed, got error: %s", result.Seeded[0].Error)
	}
	if result.Seeded[0].Rows != 20_000 {
		t.Errorf("expected 20000 seeded rows, got %d", result.Seeded[0].Rows)
	}

	if len(result.Statements) != 1 {
		t.Fatalf("expected one measured statement, got %d", len(result.Statements))
	}
	measurement := result.Statements[0]
	if measurement.StrongestLock != "AccessExclusiveLock" {
		t.Errorf("expected ACCESS EXCLUSIVE to be observed, got %q from %+v", measurement.StrongestLock, measurement.Locks)
	}
	if !lockedRelation(measurement.Locks, "events") || !lockedRelation(measurement.Locks, "events_2026") {
		t.Errorf("expected the parent and its partition to be locked, got %+v", measurement.Locks)
	}

	if len(result.Probes) != 1 || result.Probes[0].Table != "events_2026" {
		t.Fatalf("expected one reader probe on the seeded partition, got %+v", result.Probes)
	}
	if result.Probes[0].Samples == 0 {
		t.Error("expected the reader probe to have read at least once")
	}
	if result.Retry != models.RetryNotApplicable {
		t.Errorf("expected no retry question for a completed migration, got %s", result.Retry)
	}
}

func TestAnalyseReportsAFailureLoopWhenTheDeadlineIsTooShort(t *testing.T) {
	result, err := Analyse(context.Background(), Request{
		Migrations: partitionedTimeline(),
		Target:     "V3__widen_note",
		Rows:       20_000,
		Deadline:   time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Analyse failed: %v", err)
	}

	if result.Verdict != models.RuntimeExceedsDeadline {
		t.Fatalf("expected the rewrite to miss a 1 ms deadline, got %s (%s)", result.Verdict, result.Message)
	}
	if result.Retry != models.RetryFailureLoop {
		t.Errorf("expected a cancelled rewrite to leave a failure loop, got %s", result.Retry)
	}
	if !hasFinding(result.Findings, models.FindingExceedsDeadline) {
		t.Errorf("expected an EXCEEDS_DEADLINE finding, got %+v", result.Findings)
	}
}

func TestAnalyseRejectsAnUnknownTarget(t *testing.T) {
	_, err := Analyse(context.Background(), Request{Migrations: partitionedTimeline(), Target: "V9__ghost"})

	if err != models.ErrNotFound {
		t.Errorf("expected ErrNotFound for a migration outside the timeline, got %v", err)
	}
}

func lockedRelation(locks []models.LockObservation, relation string) bool {
	for _, lock := range locks {
		if lock.Relation == relation {
			return true
		}
	}
	return false
}

func hasFinding(findings []models.RuntimeFinding, findingType string) bool {
	for _, finding := range findings {
		if finding.Type == findingType {
			return true
		}
	}
	return false
}
