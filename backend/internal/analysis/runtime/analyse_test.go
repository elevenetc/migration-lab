package runtime

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"migration-lab/backend/internal/models"
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

	if !hasFinding(result.Findings, models.FindingExclusiveLock) {
		t.Errorf("expected an exclusive lock finding for the rewrite, got %+v", result.Findings)
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

// One migration per performance class, measured in one run: the observed class
// comes from counts — a changed relfilenode, rows read inside the transaction —
// so a faster or slower machine cannot move any of these answers.
func classTimeline() []models.MigrationInfo {
	return []models.MigrationInfo{
		{ID: "V1__create_accounts", Timestamp: 1, SQL: `CREATE TABLE accounts (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255),
			age INT,
			note TEXT
		);`},
		{ID: "V2__classes", Timestamp: 2, SQL: `ALTER TABLE accounts ADD COLUMN c1 TEXT;
			ALTER TABLE accounts ADD COLUMN c2 TIMESTAMP DEFAULT now();
			ALTER TABLE accounts ALTER COLUMN email TYPE TEXT;
			ALTER TABLE accounts ALTER COLUMN note SET NOT NULL;
			ALTER TABLE accounts ADD COLUMN token INT DEFAULT random()::int;
			ALTER TABLE accounts ALTER COLUMN age TYPE BIGINT;`},
	}
}

func TestAnalyseObservesThePerformanceClassOfEveryStatement(t *testing.T) {
	result, err := Analyse(context.Background(), Request{
		Migrations: classTimeline(),
		Rows:       20_000,
		Deadline:   2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("Analyse failed: %v", err)
	}
	if result.Verdict != models.RuntimeCompleted {
		t.Fatalf("expected the migration to complete inside two minutes, got %s (%s)", result.Verdict, result.Message)
	}

	want := []models.PerformanceClass{
		models.MetadataOnly, // ADD COLUMN c1 TEXT
		models.MetadataOnly, // ADD COLUMN c2 ... DEFAULT now(), stored in the catalog since PG 11
		models.MetadataOnly, // varchar(255) -> text is binary-coercible and drops the limit
		models.DataScanning, // SET NOT NULL reads every row to verify it
		models.TableRewrite, // DEFAULT random()::int is volatile however the cast hides it
		models.TableRewrite, // int -> bigint changes the on-disk representation
	}
	if len(result.Statements) != len(want) {
		t.Fatalf("expected %d measured statements, got %d", len(want), len(result.Statements))
	}

	for i, expected := range want {
		measurement := result.Statements[i]
		if measurement.ObservedClass != expected {
			t.Errorf("statement %d (%s) observed %q, want %q — read %d rows, rewrote %v",
				i, measurement.SQL, measurement.ObservedClass, expected,
				measurement.TuplesRead, measurement.RewrittenRelations)
		}
		if measurement.PredictedClass != expected {
			t.Errorf("statement %d (%s) predicted %q, want %q",
				i, measurement.SQL, measurement.PredictedClass, expected)
		}
	}

	for _, i := range []int{4, 5} {
		if !slices.Contains(result.Statements[i].RewrittenRelations, "accounts") {
			t.Errorf("statement %d should have named the rewritten relation, got %v",
				i, result.Statements[i].RewrittenRelations)
		}
	}

	if hasFinding(result.Findings, models.FindingClassUnderstated) ||
		hasFinding(result.Findings, models.FindingClassOverstated) {
		t.Errorf("expected the prediction to agree with the measurement throughout, got %+v", result.Findings)
	}
}

func TestAnalyseRejectsAnUnknownTarget(t *testing.T) {
	_, err := Analyse(context.Background(), Request{Migrations: partitionedTimeline(), Target: "V9__ghost"})

	if !errors.Is(err, models.ErrNotFound) {
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
