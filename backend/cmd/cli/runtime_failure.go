package main

import (
	"fmt"

	"migration-lab/backend/internal/models"
)

// runtimeFailure is the error that fails the process after a runtime pass. A
// target that does not complete inside the deadline fails the CI job it runs in,
// which is the point of the pass; measuring every migration only reports the
// verdicts, since an old migration must not fail every later run.
func runtimeFailure(result *models.RuntimeAnalysisResult, targetMigrationID string) error {
	if result == nil {
		return nil
	}
	if len(result.Migrations) == 0 {
		return fmt.Errorf("runtime analysis could not start: %s", result.Message)
	}
	if targetMigrationID == "" {
		return nil
	}
	for _, migration := range result.Migrations {
		if migration.MigrationID == targetMigrationID && migration.Verdict != models.RuntimeCompleted {
			return fmt.Errorf("runtime analysis of %s: %s (%s)", migration.MigrationID, migration.Verdict, migration.Message)
		}
	}
	return nil
}
