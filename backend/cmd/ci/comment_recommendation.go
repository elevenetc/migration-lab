package main

import "migration-lab/backend/internal/models"

type commentAdvice struct {
	Message    string
	NoConcerns bool
}

// This is review advice for the measured migration, not a new CI failure policy.
// Missing evidence cannot produce a clean recommendation, and timing alone does
// not establish safety: even a fast scan or rewrite scales with table size.
func commentRecommendation(migration commentMigration) commentAdvice {
	result := migration.Runtime
	if migration.Problem != "" || result == nil {
		return commentAdvice{Message: "Analysis incomplete — rerun before merging."}
	}
	if result.Retry == models.RetryManualCleanup || result.Retry == models.RetryFailureLoop {
		return commentAdvice{Message: "Do not merge — the migration cannot be retried without intervention."}
	}
	switch result.Verdict {
	case models.RuntimeFailed:
		return commentAdvice{Message: "Do not merge — the migration failed during execution."}
	case models.RuntimeExceedsDeadline:
		return commentAdvice{Message: "Do not merge — the migration exceeded the runtime limit."}
	case models.RuntimeCompleted:
	default:
		return commentAdvice{Message: "Analysis incomplete — rerun before merging."}
	}
	if migration.ExitCode != 0 {
		return commentAdvice{Message: "Analysis incomplete — resolve the analysis failure and rerun before merging."}
	}
	for _, finding := range result.Findings {
		switch finding.Type {
		case models.FindingStatementFailed, models.FindingExceedsDeadline, models.FindingInvalidIndexLeft:
			return commentAdvice{Message: "Do not merge — resolve the reported migration failure."}
		}
	}
	for _, table := range result.Seeded {
		if table.Error != "" || table.Rows == 0 {
			return commentAdvice{Message: "Analysis incomplete — resolve seeding coverage and rerun before merging."}
		}
	}
	for _, finding := range result.Findings {
		if finding.Type == models.FindingSeedFailed {
			return commentAdvice{Message: "Analysis incomplete — resolve seeding coverage and rerun before merging."}
		}
	}
	for _, finding := range result.Findings {
		switch finding.Type {
		case models.FindingExclusiveLock:
			return commentAdvice{Message: "Review before merging — reader-blocking lock duration grows with table size."}
		case models.FindingClassOverstated:
			// A pessimistic prediction is informative, not a runtime risk.
		default:
			return commentAdvice{Message: "Review before merging — runtime findings need attention."}
		}
	}
	classes := commentPerformance(result.Statements)
	if !classes.ObservedComplete {
		return commentAdvice{Message: "Review before merging — performance could not be verified for every statement."}
	}
	switch classes.Observed {
	case models.TableRewrite:
		return commentAdvice{Message: "Review before merging — the migration rewrites table data."}
	case models.DataScanning:
		return commentAdvice{Message: "Review before merging — the migration scans table data."}
	default:
		return commentAdvice{Message: "No runtime concerns found.", NoConcerns: true}
	}
}
