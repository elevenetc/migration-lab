package runtime

import (
	"context"
	"errors"
	"slices"
	"strconv"
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
	analysis, err := Analyse(context.Background(), Request{
		Migrations:        partitionedTimeline(),
		TargetMigrationID: "V3__widen_note",
		Rows:              20_000,
		Deadline:          2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("Analyse failed: %v", err)
	}
	result := onlyMigration(t, analysis)

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
	analysis, err := Analyse(context.Background(), Request{
		Migrations:        partitionedTimeline(),
		TargetMigrationID: "V3__widen_note",
		Rows:              20_000,
		Deadline:          time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Analyse failed: %v", err)
	}
	result := onlyMigration(t, analysis)

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
	analysis, err := Analyse(context.Background(), Request{
		Migrations:        classTimeline(),
		TargetMigrationID: "V2__classes",
		Rows:              20_000,
		Deadline:          2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("Analyse failed: %v", err)
	}
	result := onlyMigration(t, analysis)
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

func TestAnalyseSeedsATableWhoseOnlyColumnIsAnIdentity(t *testing.T) {
	analysis, err := Analyse(context.Background(), Request{
		Migrations: []models.MigrationInfo{
			{ID: "V1__create_accounts", Timestamp: 1, SQL: `CREATE TABLE accounts (id BIGINT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY);`},
			{ID: "V2__add_public_id", Timestamp: 2, SQL: `ALTER TABLE accounts ADD COLUMN public_id UUID DEFAULT gen_random_uuid();`},
		},
		TargetMigrationID: "V2__add_public_id",
		Rows:              1_000,
		Deadline:          2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("Analyse failed: %v", err)
	}
	result := onlyMigration(t, analysis)

	if len(result.Seeded) != 1 || result.Seeded[0].Table != "accounts" {
		t.Fatalf("expected accounts to be seeded, got %+v", result.Seeded)
	}
	if result.Seeded[0].Error != "" || result.Seeded[0].Rows != 1_000 {
		t.Errorf("expected 1000 seeded rows without error, got %+v", result.Seeded[0])
	}
	if hasFinding(result.Findings, models.FindingSeedFailed) {
		t.Errorf("expected no seed failure, got %+v", result.Findings)
	}
}

func TestAnalyseRejectsAnUnknownTarget(t *testing.T) {
	_, err := Analyse(context.Background(), Request{Migrations: partitionedTimeline(), TargetMigrationID: "V9__ghost"})

	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound for a migration outside the timeline, got %v", err)
	}
}

func accountsTimeline(last string) []models.MigrationInfo {
	return []models.MigrationInfo{
		{ID: "V1__create_accounts", Timestamp: 1, SQL: `CREATE TABLE accounts (id BIGINT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY, name TEXT);`},
		{ID: "V2__add_public_id", Timestamp: 2, SQL: `ALTER TABLE accounts ADD COLUMN public_id UUID DEFAULT gen_random_uuid();`},
		{ID: "V3__last", Timestamp: 3, SQL: last},
	}
}

func TestAnalyseEveryMeasuresEveryMigrationAgainstWhatItsPredecessorsLeft(t *testing.T) {
	result, err := Analyse(context.Background(), Request{
		Migrations: accountsTimeline(`ALTER TABLE accounts ALTER COLUMN public_id SET NOT NULL;`),
		Rows:       2_000,
		Deadline:   2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("Analyse failed: %v", err)
	}

	if result.Stopped != "" || len(result.Migrations) != 3 {
		t.Fatalf("expected all 3 migrations measured, got stopped=%q message=%q", result.Stopped, result.Message)
	}
	for _, migration := range result.Migrations {
		if migration.Verdict != models.RuntimeCompleted {
			t.Errorf("%s: expected COMPLETED, got %s (%s)", migration.MigrationID, migration.Verdict, migration.Message)
		}
	}

	if created := result.Migrations[0]; len(created.Seeded) != 0 {
		t.Errorf("expected nothing seeded before the table exists, got %+v", created.Seeded)
	}
	added := result.Migrations[1]
	if len(added.Seeded) != 1 || added.Seeded[0].Rows != 2_000 {
		t.Errorf("expected accounts seeded to 2000 rows, got %+v", added.Seeded)
	}
	if len(added.Statements) != 1 || added.Statements[0].ObservedClass != models.TableRewrite {
		t.Errorf("expected the volatile default to rewrite the seeded table, got %+v", added.Statements)
	}
	// The table was seeded for V2 already, so V3 finds it full rather than growing it.
	if last := result.Migrations[2]; len(last.Seeded) != 1 || last.Seeded[0].Rows != 2_000 {
		t.Errorf("expected accounts to stay at 2000 rows, got %+v", last.Seeded)
	}
}

func TestAnalyseEveryForcesAMigrationThroughWhenTheDeadlineIsTooShort(t *testing.T) {
	result, err := Analyse(context.Background(), Request{
		Migrations: accountsTimeline(`ALTER TABLE accounts ALTER COLUMN public_id SET NOT NULL;`),
		Rows:       20_000,
		Deadline:   time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Analyse failed: %v", err)
	}

	if result.Stopped != "" || len(result.Migrations) != 3 {
		t.Fatalf("expected the run to reach the end, got stopped=%q message=%q", result.Stopped, result.Message)
	}
	if verdict := result.Migrations[1].Verdict; verdict != models.RuntimeExceedsDeadline {
		t.Errorf("expected the rewrite to exceed a 1 ms deadline, got %s", verdict)
	}
	// V3 only runs when V2 was forced through, since it needs the column V2 adds.
	if last := result.Migrations[2]; last.Verdict == models.RuntimeFailed {
		t.Errorf("expected V3 to find public_id, got %s", last.Message)
	}
}

func TestAnalyseEveryStopsAtAMigrationThatFailsWithoutADeadline(t *testing.T) {
	result, err := Analyse(context.Background(), Request{
		Migrations: []models.MigrationInfo{
			{ID: "V1__create_accounts", Timestamp: 1, SQL: `CREATE TABLE accounts (id BIGINT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY, name TEXT);`},
			{ID: "V2__require_note", Timestamp: 2, SQL: `ALTER TABLE accounts ADD COLUMN note TEXT NOT NULL;`},
			{ID: "V3__add_age", Timestamp: 3, SQL: `ALTER TABLE accounts ADD COLUMN age INT;`},
		},
		Rows:     100,
		Deadline: 2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("Analyse failed: %v", err)
	}

	if result.Stopped != "V2__require_note" {
		t.Errorf("expected the run to stop at V2, got stopped=%q message=%q", result.Stopped, result.Message)
	}
	if len(result.Migrations) != 2 || result.Migrations[1].Verdict != models.RuntimeFailed {
		t.Errorf("expected V2 measured as FAILED and V3 left out, got %+v", result.Migrations)
	}
}

func TestAnalyseEveryRejectsAnEmptyTimeline(t *testing.T) {
	if _, err := Analyse(context.Background(), Request{}); !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound for an empty timeline, got %v", err)
	}
}

func TestAnalyseMeasuresOnlyTheTargetAndRollsItBack(t *testing.T) {
	analysis, err := Analyse(context.Background(), Request{
		Migrations:        accountsTimeline(`ALTER TABLE accounts ALTER COLUMN public_id SET NOT NULL;`),
		TargetMigrationID: "V2__add_public_id",
		Rows:              1_000,
		Deadline:          2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("Analyse failed: %v", err)
	}

	if len(analysis.Migrations) != 1 || analysis.Migrations[0].MigrationID != "V2__add_public_id" {
		t.Fatalf("expected only the target measured, got %+v", analysis.Migrations)
	}
	if analysis.Message != "all 1 migrations completed in "+strconv.FormatInt(analysis.Migrations[0].Statements[0].DurationMs, 10)+" ms" {
		t.Errorf("unexpected message %q", analysis.Message)
	}
}

// onlyMigration is the one measured migration of a targeted analysis.
func onlyMigration(t *testing.T, analysis models.RuntimeAnalysisResult) models.RuntimeMigrationResult {
	t.Helper()
	if len(analysis.Migrations) != 1 {
		t.Fatalf("expected exactly one measured migration, got %d: %s", len(analysis.Migrations), analysis.Message)
	}
	return analysis.Migrations[0]
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
