package runtime

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"migration-lab/backend/internal/models"
)

func TestPartitionStrategy(t *testing.T) {
	cases := map[string]string{
		"RANGE (year)":         StrategyRange,
		"LIST (region)":        StrategyList,
		"HASH (id)":            StrategyHash,
		"RANGE (((a + b)))":    StrategyRange,
		"":                     "",
		"SOMETHING ELSE (col)": "",
	}

	for partkeydef, want := range cases {
		if got := PartitionStrategy(partkeydef); got != want {
			t.Errorf("PartitionStrategy(%q) = %q, want %q", partkeydef, got, want)
		}
	}
}

func TestPartitionKeyColumns(t *testing.T) {
	cases := []struct {
		partkeydef string
		want       []string
	}{
		{"RANGE (year)", []string{"year"}},
		{"RANGE (year, month)", []string{"year", "month"}},
		{`LIST ("Region")`, []string{"Region"}},
		{"RANGE (((a + b)))", nil},
		{"HASH", nil},
	}

	for _, c := range cases {
		if got := PartitionKeyColumns(c.partkeydef); !reflect.DeepEqual(got, c.want) {
			t.Errorf("PartitionKeyColumns(%q) = %v, want %v", c.partkeydef, got, c.want)
		}
	}
}

func TestPartitionSeedValues(t *testing.T) {
	cases := []struct {
		bound string
		want  []string
	}{
		{"FOR VALUES FROM (2026) TO (2027)", []string{"2026"}},
		{"FOR VALUES FROM ('2024-01-01') TO ('2025-01-01')", []string{"'2024-01-01'"}},
		{"FOR VALUES FROM (2026, 1) TO (2026, 4)", []string{"2026", "1"}},
		{"FOR VALUES IN ('eu', 'us')", []string{"'eu'"}},
		{"FOR VALUES FROM (MINVALUE) TO (2020)", nil},
		{"FOR VALUES WITH (modulus 4, remainder 0)", nil},
		{"DEFAULT", nil},
	}

	for _, c := range cases {
		if got := PartitionSeedValues(c.bound); !reflect.DeepEqual(got, c.want) {
			t.Errorf("PartitionSeedValues(%q) = %v, want %v", c.bound, got, c.want)
		}
	}
}

func TestSeedExprCoversCommonTypesAndSkipsUnknownOnes(t *testing.T) {
	if SeedExpr("integer", 0) != "g.i" {
		t.Errorf("expected integers to be seeded from the series, got %q", SeedExpr("integer", 0))
	}
	if got := SeedExpr("character varying", 12); got != "left(('row_' || g.i), 12)" {
		t.Errorf("expected a bounded varchar to be truncated, got %q", got)
	}
	if got := SeedExpr("text", 0); got != "('row_' || g.i)" {
		t.Errorf("expected unbounded text to be untruncated, got %q", got)
	}
	if got := SeedExpr("USER-DEFINED", 0); got != "" {
		t.Errorf("expected no expression for an enum column, got %q", got)
	}
}

func TestSeedStatement(t *testing.T) {
	plan := SeedPlan{Table: "events_2026", Columns: []SeedColumn{
		{Name: "year", Expr: "2026"},
		{Name: "note", Expr: "('row_' || g.i)"},
	}}

	want := `INSERT INTO "events_2026" ("year", "note") SELECT 2026, ('row_' || g.i) FROM generate_series(1, 1000) AS g(i)`
	if got := SeedStatement(plan, 1, 1000); got != want {
		t.Errorf("SeedStatement =\n%s\nwant\n%s", got, want)
	}
}

func TestSeedStatementContinuesPastTheRowsATableHolds(t *testing.T) {
	plan := SeedPlan{Table: "accounts", Columns: []SeedColumn{{Name: "id", Expr: "g.i"}}}
	want := `INSERT INTO "accounts" ("id") SELECT g.i FROM generate_series(4, 1000) AS g(i)`
	if got := SeedStatement(plan, 4, 1000); got != want {
		t.Errorf("SeedStatement =\n%s\nwant\n%s", got, want)
	}
}

func TestMissingRowsTopsATableUpWithoutEverShrinkingIt(t *testing.T) {
	cases := []struct{ existing, target, want int64 }{
		{0, 1000, 1000},
		{3, 1000, 997},
		{1000, 1000, 0},
		{1500, 1000, 0},
	}
	for _, c := range cases {
		if got := missingRows(c.existing, c.target); got != c.want {
			t.Errorf("missingRows(%d, %d) = %d, want %d", c.existing, c.target, got, c.want)
		}
	}
}

func TestSeedStatementLeavesEveryColumnToItsDefaultWithoutFillableColumns(t *testing.T) {
	want := `INSERT INTO "accounts" SELECT FROM generate_series(1, 1000) AS g(i)`
	if got := SeedStatement(SeedPlan{Table: "accounts"}, 1, 1000); got != want {
		t.Errorf("SeedStatement =\n%s\nwant\n%s", got, want)
	}
}

func intColumn(name string) CatalogColumn {
	return CatalogColumn{Name: name, DataType: "integer"}
}

func TestSeedPlansPlanAStandaloneTable(t *testing.T) {
	catalog := map[string]CatalogTable{
		"users": {Name: "users", Columns: []CatalogColumn{
			{Name: "id", DataType: "integer", AutoFilled: true},
			intColumn("age"),
		}},
	}

	plans := SeedPlans(catalog, []string{"users"})

	want := []SeedPlan{{Table: "users", Columns: []SeedColumn{{Name: "age", Expr: "g.i"}}}}
	if !reflect.DeepEqual(plans, want) {
		t.Errorf("SeedPlans = %+v, want %+v", plans, want)
	}
}

func rangePartitioned() map[string]CatalogTable {
	return map[string]CatalogTable{
		"events": {
			Name:            "events",
			Partitioned:     true,
			PartitionKeyDef: "RANGE (year)",
			Columns:         []CatalogColumn{intColumn("year"), intColumn("value")},
		},
		"events_2026": {
			Name:    "events_2026",
			Parent:  "events",
			Bound:   "FOR VALUES FROM (2026) TO (2027)",
			Columns: []CatalogColumn{intColumn("year"), intColumn("value")},
		},
		"events_2027": {
			Name:    "events_2027",
			Parent:  "events",
			Bound:   "FOR VALUES FROM (2027) TO (2028)",
			Columns: []CatalogColumn{intColumn("year"), intColumn("value")},
		},
	}
}

// The migration names the parent; the plans cross to the tables actually holding
// rows, which are its partitions.
func TestSeedPlansPinPartitionKeysInsideEachBound(t *testing.T) {
	plans := SeedPlans(rangePartitioned(), []string{"events"})

	want := []SeedPlan{
		{Table: "events_2026", Columns: []SeedColumn{{Name: "year", Expr: "2026"}, {Name: "value", Expr: "g.i"}}},
		{Table: "events_2027", Columns: []SeedColumn{{Name: "year", Expr: "2027"}, {Name: "value", Expr: "g.i"}}},
	}
	if !reflect.DeepEqual(plans, want) {
		t.Errorf("SeedPlans = %+v, want %+v", plans, want)
	}
}

func TestSeedPlansPinTheKeyOfADirectlyNamedPartition(t *testing.T) {
	plans := SeedPlans(rangePartitioned(), []string{"events_2027"})

	want := []SeedPlan{
		{Table: "events_2027", Columns: []SeedColumn{{Name: "year", Expr: "2027"}, {Name: "value", Expr: "g.i"}}},
	}
	if !reflect.DeepEqual(plans, want) {
		t.Errorf("SeedPlans = %+v, want %+v", plans, want)
	}
}

func TestSeedPlansFallBackToTheParentWhenNoBoundAdmitsAConstant(t *testing.T) {
	catalog := map[string]CatalogTable{
		"events": {
			Name:            "events",
			Partitioned:     true,
			PartitionKeyDef: "HASH (id)",
			Columns:         []CatalogColumn{intColumn("id")},
		},
		"events_0": {
			Name:    "events_0",
			Parent:  "events",
			Bound:   "FOR VALUES WITH (modulus 2, remainder 0)",
			Columns: []CatalogColumn{intColumn("id")},
		},
	}

	plans := SeedPlans(catalog, []string{"events"})

	if len(plans) != 1 || plans[0].Table != "events" {
		t.Errorf("expected the hash-partitioned parent to be seeded so routing places the rows, got %+v", plans)
	}
}

func TestSeedPlansSkipUnknownTables(t *testing.T) {
	if plans := SeedPlans(map[string]CatalogTable{}, []string{"ghost"}); plans != nil {
		t.Errorf("expected no plan for a table the catalog does not hold, got %+v", plans)
	}
}

func TestTouchedTablesSkipsTablesTheMigrationCreatesOrDrops(t *testing.T) {
	migration := &models.Migration{Statements: []models.Statement{
		{Index: 0, Operations: []models.Operation{models.CreateTable{TableName: "fresh"}}},
		{Index: 1, Operations: []models.Operation{
			models.AddColumn{TableName: "users"},
			models.SetNotNull{TableName: "users"},
			models.AlterColumnType{TableName: "orders"},
		}},
		{Index: 2, Operations: []models.Operation{models.DropTable{TableName: "legacy"}}},
	}}

	if got := TouchedTables(migration); !reflect.DeepEqual(got, []string{"users", "orders"}) {
		t.Errorf("TouchedTables = %v, want [users orders]", got)
	}
}

func TestStrongestLock(t *testing.T) {
	locks := []models.LockObservation{
		{Mode: "AccessShareLock", Relation: "users"},
		{Mode: "AccessExclusiveLock", Relation: "events"},
		{Mode: "RowExclusiveLock", Relation: "orders"},
	}

	if got := StrongestLock(locks); got != "AccessExclusiveLock" {
		t.Errorf("StrongestLock = %q, want AccessExclusiveLock", got)
	}
	if got := StrongestLock(nil); got != "" {
		t.Errorf("StrongestLock of nothing = %q, want empty", got)
	}
}

func TestBlocksReaders(t *testing.T) {
	if !BlocksReaders("AccessExclusiveLock") {
		t.Error("expected ACCESS EXCLUSIVE to block readers")
	}
	if BlocksReaders("ShareRowExclusiveLock") {
		t.Error("expected SHARE ROW EXCLUSIVE to leave readers alone")
	}
	if BlocksReaders("") {
		t.Error("expected no lock to block nothing")
	}
}

func TestTransactionalRejectsStatementsPostgresRefusesInATransaction(t *testing.T) {
	if !transactional([]string{"ALTER TABLE users ADD COLUMN a INT;"}) {
		t.Error("expected a plain ALTER to be measurable inside a transaction")
	}
	if transactional([]string{"CREATE INDEX CONCURRENTLY i ON users (a);"}) {
		t.Error("expected CREATE INDEX CONCURRENTLY to force a run without a transaction")
	}
}

func TestFindingsReportWhatStoppedTheRunAndWhichLocksItHeld(t *testing.T) {
	migration := &models.Migration{
		ID: "V3__widen_note",
		Statements: []models.Statement{{
			Index:      0,
			SQL:        "ALTER TABLE events ALTER COLUMN note TYPE VARCHAR(200)",
			Operations: []models.Operation{models.AlterColumnType{TableName: "events"}},
		}},
	}
	result := models.RuntimeMigrationResult{
		DeadlineMs: 5000,
		Seeded:     []models.SeededTable{{Table: "events_2026", Rows: 1000}},
		Statements: []models.StatementMeasurement{{
			StatementIndex: 0,
			SQL:            "ALTER TABLE events ALTER COLUMN note TYPE VARCHAR(200)",
			DurationMs:     9000,
			PredictedClass: models.TableRewrite,
			StrongestLock:  "AccessExclusiveLock",
			Locks: []models.LockObservation{
				{Mode: "AccessExclusiveLock", Relation: "events"},
				{Mode: "AccessExclusiveLock", Relation: "events_2026"},
			},
			Verdict: models.RuntimeExceedsDeadline,
		}},
		Retry: models.RetryFailureLoop,
	}

	findings := Findings(migration, result)

	types := make([]string, len(findings))
	for i, finding := range findings {
		types[i] = finding.Type
	}
	want := []string{models.FindingExceedsDeadline, models.FindingExclusiveLock}
	if !reflect.DeepEqual(types, want) {
		t.Errorf("finding types = %v, want %v", types, want)
	}
	if findings[0].TableName != "events" {
		t.Errorf("expected the deadline finding on events, got %q", findings[0].TableName)
	}
	if findings[0].OperationID.MigrationID != "V3__widen_note" || findings[0].OperationID.OpIndex != models.StatementScoped {
		t.Errorf("expected a statement-scoped id of the migration, got %+v", findings[0].OperationID)
	}
}

func TestFindingsStaySilentForAMetadataOnlyAlter(t *testing.T) {
	migration := &models.Migration{
		ID: "V2__add_note",
		Statements: []models.Statement{{
			Index:      0,
			SQL:        "ALTER TABLE accounts ADD COLUMN note TEXT",
			Operations: []models.Operation{models.AddColumn{TableName: "accounts"}},
		}},
	}
	// Every ALTER TABLE takes ACCESS EXCLUSIVE, but a metadata-only one holds it for
	// work that does not grow with the table.
	result := models.RuntimeMigrationResult{
		DeadlineMs: 5000,
		Seeded:     []models.SeededTable{{Table: "accounts", Rows: 1_000_000}},
		Statements: []models.StatementMeasurement{{
			StatementIndex: 0,
			SQL:            "ALTER TABLE accounts ADD COLUMN note TEXT",
			DurationMs:     0,
			PredictedClass: models.MetadataOnly,
			ObservedClass:  models.MetadataOnly,
			StrongestLock:  "AccessExclusiveLock",
			Locks:          []models.LockObservation{{Mode: "AccessExclusiveLock", Relation: "accounts"}},
			Verdict:        models.RuntimeCompleted,
		}},
		Retry: models.RetryNotApplicable,
	}

	if findings := Findings(migration, result); len(findings) != 0 {
		t.Errorf("expected no finding for a metadata-only alter, got %+v", findings)
	}
}

// The gate is the statement's performance class, not its duration: a scan is
// worth reporting however fast the machine that measured it happened to be.
func TestFindingsReportAScanningLockHoweverFastThisMachineWas(t *testing.T) {
	migration := &models.Migration{
		ID: "V5__note_not_null",
		Statements: []models.Statement{{
			Index:      0,
			SQL:        "ALTER TABLE accounts ALTER COLUMN note SET NOT NULL",
			Operations: []models.Operation{models.SetNotNull{TableName: "accounts", ColumnName: "note"}},
		}},
	}
	result := models.RuntimeMigrationResult{
		Statements: []models.StatementMeasurement{{
			StatementIndex: 0,
			SQL:            "ALTER TABLE accounts ALTER COLUMN note SET NOT NULL",
			DurationMs:     3,
			PredictedClass: models.DataScanning,
			ObservedClass:  models.DataScanning,
			StrongestLock:  "AccessExclusiveLock",
			Locks:          []models.LockObservation{{Mode: "AccessExclusiveLock", Relation: "accounts"}},
			Verdict:        models.RuntimeCompleted,
		}},
		Retry: models.RetryNotApplicable,
	}

	findings := Findings(migration, result)

	if len(findings) != 1 || findings[0].Type != models.FindingExclusiveLock {
		t.Errorf("expected a data-scanning lock to be reported at any duration, got %+v", findings)
	}
}

func TestFindingsReportALockOfAStatementStaticAnalysisDoesNotModel(t *testing.T) {
	migration := &models.Migration{ID: "V9__vacuum", Statements: []models.Statement{}}
	result := models.RuntimeMigrationResult{
		Statements: []models.StatementMeasurement{{
			StatementIndex: 0,
			SQL:            "VACUUM FULL accounts",
			StrongestLock:  "AccessExclusiveLock",
			Locks:          []models.LockObservation{{Mode: "AccessExclusiveLock", Relation: "accounts"}},
			Verdict:        models.RuntimeCompleted,
		}},
		Retry: models.RetryNotApplicable,
	}

	findings := Findings(migration, result)

	if len(findings) != 1 || findings[0].Type != models.FindingExclusiveLock {
		t.Errorf("expected an unmodelled statement's lock to be reported, got %+v", findings)
	}
}

func TestFindingsReportATableThatCouldNotBeSeeded(t *testing.T) {
	migration := &models.Migration{ID: "V2__alter"}
	result := models.RuntimeMigrationResult{
		Seeded: []models.SeededTable{{Table: "orders", Error: "invalid input syntax"}},
		Retry:  models.RetryNotApplicable,
	}

	findings := Findings(migration, result)

	if len(findings) != 1 || findings[0].Type != models.FindingSeedFailed {
		t.Fatalf("expected one SEED_FAILED finding, got %+v", findings)
	}
	if findings[0].TableName != "orders" {
		t.Errorf("expected the finding on orders, got %q", findings[0].TableName)
	}
}

func TestResumeIndexRestartsARolledBackMigrationAndResumesAnyOther(t *testing.T) {
	stoppedAtSecond := models.RuntimeMigrationResult{Statements: []models.StatementMeasurement{
		{StatementIndex: 0, Verdict: models.RuntimeCompleted},
		{StatementIndex: 1, Verdict: models.RuntimeExceedsDeadline},
	}}
	transactional := []string{"ALTER TABLE a ADD COLUMN b INT", "ALTER TABLE a ADD COLUMN c INT"}
	concurrent := []string{"CREATE INDEX CONCURRENTLY a_b ON a (b)", "CREATE INDEX CONCURRENTLY a_c ON a (c)"}

	if got := resumeIndex(transactional, stoppedAtSecond); got != 0 {
		t.Errorf("rolled-back migration resumes at %d, want 0", got)
	}
	if got := resumeIndex(concurrent, stoppedAtSecond); got != 1 {
		t.Errorf("non-transactional migration resumes at %d, want 1", got)
	}
	if got := resumeIndex(concurrent, models.RuntimeMigrationResult{}); got != 0 {
		t.Errorf("unmeasured migration resumes at %d, want 0", got)
	}
}

func TestSplitMeasuresTheTargetAloneOrTheWholeTimeline(t *testing.T) {
	v1, v2, v3 := models.MigrationInfo{ID: "v1"}, models.MigrationInfo{ID: "v2"}, models.MigrationInfo{ID: "v3"}
	timeline := []models.MigrationInfo{v1, v2, v3}

	prior, toMeasure, err := split(timeline, "")
	if err != nil || len(prior) != 0 || !reflect.DeepEqual(toMeasure, timeline) {
		t.Errorf("split with no target = %v, %v (%v), want none prior and every one measured", prior, toMeasure, err)
	}
	prior, toMeasure, err = split(timeline, "v2")
	if err != nil || !reflect.DeepEqual(prior, []models.MigrationInfo{v1}) || !reflect.DeepEqual(toMeasure, []models.MigrationInfo{v2}) {
		t.Errorf("split(v2) = %v, %v (%v), want v1 prior and v2 measured", prior, toMeasure, err)
	}
}

func TestSplitRejectsAnEmptyOrMismatchedTimeline(t *testing.T) {
	if _, _, err := split(nil, ""); !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound for an empty timeline, got %v", err)
	}
	if _, _, err := split([]models.MigrationInfo{{ID: "v1"}}, "v9"); !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound for a migration outside the timeline, got %v", err)
	}
}

func TestRowsOfFallsBackToTheDefault(t *testing.T) {
	if got := rowsOf(0); got != DefaultRows {
		t.Errorf("rows without a request = %d, want %d", got, DefaultRows)
	}
	if got := rowsOf(250); got != 250 {
		t.Errorf("rows with a request = %d, want 250", got)
	}
}

func TestDeadlineOfFallsBackToTheDefault(t *testing.T) {
	if got := deadlineOf(0); got != DefaultDeadline {
		t.Errorf("deadline without a request = %v, want %v", got, DefaultDeadline)
	}
	if got := deadlineOf(time.Second); got != time.Second {
		t.Errorf("deadline with a request = %v, want 1s", got)
	}
}

func TestAnalysisMessageCountsTheMigrationsThatDidNotComplete(t *testing.T) {
	completed := models.RuntimeMigrationResult{Verdict: models.RuntimeCompleted, Statements: []models.StatementMeasurement{
		{DurationMs: 120, Verdict: models.RuntimeCompleted},
	}}
	cancelled := models.RuntimeMigrationResult{Verdict: models.RuntimeExceedsDeadline, Statements: []models.StatementMeasurement{
		{DurationMs: 5000, Verdict: models.RuntimeExceedsDeadline},
	}}

	if got := analysisMessage([]models.RuntimeMigrationResult{completed, completed}); got != "all 2 migrations completed in 240 ms" {
		t.Errorf("message = %q", got)
	}
	if got := analysisMessage([]models.RuntimeMigrationResult{completed, cancelled}); got != "1 of 2 migrations did not complete" {
		t.Errorf("message = %q", got)
	}
}

func TestMessageNamesTheStatementThatEndedTheRun(t *testing.T) {
	completed := models.RuntimeMigrationResult{Statements: []models.StatementMeasurement{
		{StatementIndex: 0, DurationMs: 120, Verdict: models.RuntimeCompleted},
		{StatementIndex: 1, DurationMs: 80, Verdict: models.RuntimeCompleted},
	}}
	if got := message(completed); got != "all 2 statements completed in 200 ms" {
		t.Errorf("message = %q", got)
	}

	cancelled := models.RuntimeMigrationResult{DeadlineMs: 5000, Statements: []models.StatementMeasurement{
		{StatementIndex: 1, Verdict: models.RuntimeExceedsDeadline},
	}}
	if got := message(cancelled); got != "statement 1 was cancelled after the 5000 ms deadline" {
		t.Errorf("message = %q", got)
	}

	if got := message(models.RuntimeMigrationResult{}); got != "the migration has no statements to run" {
		t.Errorf("message = %q", got)
	}
}
