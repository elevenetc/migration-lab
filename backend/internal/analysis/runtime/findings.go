package runtime

import (
	"fmt"
	"strings"

	"migration-lab/backend/internal/models"
)

// Findings turns the measurements of a run into warning-shaped findings: first
// what stopped the migration, then what it held up while it ran, then what could
// not be measured because a table stayed empty.
func Findings(migration *models.Migration, result models.RuntimeAnalysisResult) []models.RuntimeFinding {
	// Empty rather than nil: a clean run must marshal as [] for the frontend's
	// RuntimeFinding[], and a clean run is the common case.
	findings := []models.RuntimeFinding{}

	for _, measurement := range result.Statements {
		findings = append(findings, statementFindings(migration, result, measurement)...)
	}

	for _, probe := range result.Probes {
		if probe.BlockedMs <= 0 {
			continue
		}
		findings = append(findings, models.RuntimeFinding{
			Type:        models.FindingBlocksReaders,
			OperationID: statementID(migration, models.StatementScoped),
			TableName:   probe.Table,
			Message: fmt.Sprintf(
				"a concurrent reader of %s was waiting on a lock the migration held for %d ms of the run, "+
					"its slowest read taking %d ms",
				probe.Table, probe.BlockedMs, probe.MaxLatencyMs),
		})
	}

	for _, seeded := range result.Seeded {
		if seeded.Error == "" {
			continue
		}
		findings = append(findings, models.RuntimeFinding{
			Type:        models.FindingSeedFailed,
			OperationID: statementID(migration, models.StatementScoped),
			TableName:   seeded.Table,
			Message: fmt.Sprintf("%s could not be seeded, so its measurements are of an empty table: %s",
				seeded.Table, seeded.Error),
		})
	}

	if result.Retry == models.RetryManualCleanup {
		findings = append(findings, models.RuntimeFinding{
			Type:        models.FindingInvalidIndexLeft,
			OperationID: statementID(migration, models.StatementScoped),
			Message: "the cancelled run left an invalid index behind; the next attempt fails until it is " +
				"dropped by hand, so a restarting pod never converges",
		})
	}

	return findings
}

func statementFindings(migration *models.Migration, result models.RuntimeAnalysisResult, measurement models.StatementMeasurement) []models.RuntimeFinding {
	table := tableOfSQL(migration, measurement.SQL)
	id := statementID(migration, measurement.StatementIndex)

	var findings []models.RuntimeFinding

	switch measurement.Verdict {
	case models.RuntimeExceedsDeadline:
		findings = append(findings, models.RuntimeFinding{
			Type:        models.FindingExceedsDeadline,
			OperationID: id,
			TableName:   table,
			Message: fmt.Sprintf(
				"statement %d was still running after the %d ms deadline and was cancelled; a pod killed at "+
					"that point restarts the migration from the beginning",
				measurement.StatementIndex, result.DeadlineMs),
		})
	case models.RuntimeFailed:
		findings = append(findings, models.RuntimeFinding{
			Type:        models.FindingStatementFailed,
			OperationID: id,
			TableName:   table,
			Message:     fmt.Sprintf("statement %d failed: %s", measurement.StatementIndex, measurement.Error),
		})
	}

	findings = append(findings, mismatchFindings(id, table, measurement)...)

	if holdsALockThatScales(measurement) {
		findings = append(findings, models.RuntimeFinding{
			Type:        models.FindingExclusiveLock,
			OperationID: id,
			TableName:   table,
			Message: fmt.Sprintf(
				"statement %d held %s on %s for %d ms, and its cost scales with table size, so the hold grows with the table",
				measurement.StatementIndex, measurement.StrongestLock,
				lockedRelations(measurement.Locks), measurement.DurationMs),
		})
	}

	return findings
}

// mismatchFindings reports the prediction disagreeing with the measurement, which
// is what tells a reader how much the offline classifier can be trusted. Silent
// unless both classes are known: an unmeasured statement is not evidence.
func mismatchFindings(id models.OperationID, table string, measurement models.StatementMeasurement) []models.RuntimeFinding {
	observed, predicted := measurement.ObservedClass, measurement.PredictedClass
	if observed == "" || predicted == "" || observed == predicted {
		return nil
	}

	if models.WorseThan(observed, predicted) {
		return []models.RuntimeFinding{{
			Type:        models.FindingClassUnderstated,
			OperationID: id,
			TableName:   table,
			Message: fmt.Sprintf(
				"static analysis predicted %s for statement %d, and it measured as %s: %s. "+
					"A migration like this passes the offline gate and does the work in production",
				predicted, measurement.StatementIndex, observed, whatItDid(measurement)),
		}}
	}

	return []models.RuntimeFinding{{
		Type:        models.FindingClassOverstated,
		OperationID: id,
		TableName:   table,
		Message: fmt.Sprintf(
			"static analysis predicted %s for statement %d, and it measured as %s: %s. "+
				"The prediction is pessimistic here, not the migration risky",
			predicted, measurement.StatementIndex, observed, whatItDid(measurement)),
	}}
}

// whatItDid names the counts the observed class was derived from, so the finding
// carries its own evidence.
func whatItDid(measurement models.StatementMeasurement) string {
	if len(measurement.RewrittenRelations) > 0 {
		return fmt.Sprintf("it rewrote %s and read %d rows",
			strings.Join(measurement.RewrittenRelations, ", "), measurement.TuplesRead)
	}
	if measurement.TuplesRead > 0 {
		return fmt.Sprintf("it read %d rows and rewrote nothing", measurement.TuplesRead)
	}
	return "it read no rows and rewrote nothing"
}

// holdsALockThatScales decides whether a reader-blocking lock is worth reporting,
// and asks what the statement cost rather than a stopwatch.
//
// Every ALTER TABLE takes ACCESS EXCLUSIVE, the metadata-only ones included, so
// the mode alone is not a finding. What separates a harmless hold from a
// dangerous one is not how many milliseconds it took on this machine — CI
// hardware would move that number, and a fast enough runner would hide it — but
// whether the work under the lock is constant in table size or grows with it.
//
// The measured class answers that where it is usable, and the predicted one
// stands in where it is not. A statement neither measured nor modelled (a VACUUM
// FULL, an index build) is reported on the strength of the observation alone:
// nothing said it was safe.
func holdsALockThatScales(measurement models.StatementMeasurement) bool {
	if !BlocksReaders(measurement.StrongestLock) {
		return false
	}
	if class := effectiveClass(measurement); class != "" {
		return class != models.MetadataOnly
	}
	return true
}

// effectiveClass is the class to reason about: what was measured, or what was
// predicted when nothing usable was measured.
func effectiveClass(measurement models.StatementMeasurement) models.PerformanceClass {
	if measurement.ObservedClass != "" {
		return measurement.ObservedClass
	}
	return measurement.PredictedClass
}

// classOfSQL is the performance class of the parsed statement with this SQL, and
// whether static analysis modelled it at all.
func classOfSQL(migration *models.Migration, sql string) (models.PerformanceClass, bool) {
	for _, statement := range migration.Statements {
		if statement.SQL == sql {
			return models.WorstPerformanceClass(statement.Operations), true
		}
	}
	return models.MetadataOnly, false
}

// lockedRelations names the relations a statement locked, so a fan-out over
// partitions is visible rather than summarised as one table.
func lockedRelations(locks []models.LockObservation) string {
	var relations []string
	seen := map[string]bool{}
	for _, lock := range locks {
		if BlocksReaders(lock.Mode) && !seen[lock.Relation] {
			seen[lock.Relation] = true
			relations = append(relations, lock.Relation)
		}
	}
	return strings.Join(relations, ", ")
}

// tableOfSQL is the table the parsed statement with this SQL operates on, "" for
// a statement the parser does not model.
func tableOfSQL(migration *models.Migration, sql string) string {
	for _, statement := range migration.Statements {
		if statement.SQL != sql {
			continue
		}
		for _, operation := range statement.Operations {
			if table := operationTable(operation); table != "" {
				return table
			}
		}
	}
	return ""
}

func statementID(migration *models.Migration, statementIndex int) models.OperationID {
	return models.OperationID{
		MigrationID:    migration.ID,
		StatementIndex: statementIndex,
		OpIndex:        models.StatementScoped,
	}
}
