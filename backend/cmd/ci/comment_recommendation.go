package main

import "migration-lab/backend/internal/models"

// This is review advice for the measured migration, not a new CI failure policy.
// Missing evidence cannot produce a clean recommendation, and timing alone does
// not establish safety: even a fast scan or rewrite scales with table size.
func commentRecommendation(migration commentMigration) string {
	result := migration.Runtime
	if migration.Problem != "" || result == nil {
		return "Analysis incomplete — rerun before merging."
	}
	if result.Retry == models.RetryManualCleanup || result.Retry == models.RetryFailureLoop {
		return "Do not merge — the migration cannot be retried without intervention."
	}
	switch result.Verdict {
	case models.RuntimeFailed:
		return "Do not merge — the migration failed during execution."
	case models.RuntimeExceedsDeadline:
		return "Do not merge — the migration exceeded the runtime limit."
	case models.RuntimeCompleted:
	default:
		return "Analysis incomplete — rerun before merging."
	}
	if migration.ExitCode != 0 {
		return "Analysis incomplete — resolve the analysis failure and rerun before merging."
	}
	for _, finding := range result.Findings {
		switch finding.Type {
		case models.FindingStatementFailed, models.FindingExceedsDeadline, models.FindingInvalidIndexLeft:
			return "Do not merge — resolve the reported migration failure."
		}
	}
	for _, table := range result.Seeded {
		if table.Error != "" || table.Rows == 0 {
			return "Analysis incomplete — resolve seeding coverage and rerun before merging."
		}
	}
	for _, finding := range result.Findings {
		if finding.Type == models.FindingSeedFailed {
			return "Analysis incomplete — resolve seeding coverage and rerun before merging."
		}
	}
	for _, finding := range result.Findings {
		switch finding.Type {
		case models.FindingBlocksReaders:
			return "Review before merging — concurrent readers were blocked."
		case models.FindingExclusiveLock:
			return "Review before merging — reader-blocking lock duration grows with table size."
		case models.FindingClassOverstated:
			// A pessimistic prediction is informative, not a runtime risk.
		default:
			return "Review before merging — runtime findings need attention."
		}
	}
	classes := commentPerformance(result.Statements)
	if !classes.ObservedComplete {
		return "Review before merging — performance could not be verified for every statement."
	}
	switch classes.Observed {
	case models.TableRewrite:
		return "Review before merging — the migration rewrites table data."
	case models.DataScanning:
		return "Review before merging — the migration scans table data."
	default:
		return "No runtime concerns found."
	}
}
