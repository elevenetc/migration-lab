package main

import (
	"fmt"
	"os"
	"path/filepath"

	"migration-lab/backend/internal/models"
)

type commentSummary struct {
	Problem         string
	ExistingChanges int
	Migrations      []commentMigration
}

type commentMigration struct {
	Name     string
	Problem  string
	ExitCode int
	Runtime  *models.RuntimeAnalysisResult
}

// Missing or partial artifacts become visible diagnostics instead of leaving an
// earlier successful comment on the PR after a failed run.
func loadCommentSummary(output, head string) commentSummary {
	var plan migrationPlan
	if err := readJSON(filepath.Join(output, "plan.json"), &plan); err != nil {
		return commentSummary{Problem: "The migration selection plan is unavailable. Check the workflow logs."}
	}
	if plan.Head != head {
		return commentSummary{Problem: "The selection plan does not match the analyzed commit."}
	}
	summary := commentSummary{ExistingChanges: len(plan.ExistingChanges)}
	if data, err := os.ReadFile(filepath.Join(output, "automation-error.txt")); err == nil {
		summary.Problem = string(data)
	}
	if len(plan.Targets) == 0 {
		return summary
	}
	var results []migrationResult
	resultsErr := readJSON(filepath.Join(output, "results.json"), &results)
	byName := make(map[string]migrationResult, len(results))
	for _, result := range results {
		byName[result.Migration] = result
	}
	for _, target := range plan.Targets {
		migration := commentMigration{Name: target}
		result, found := byName[target]
		switch {
		case resultsErr != nil || !found:
			migration.Problem = "No runtime result was collected. Check the workflow logs."
		case filepath.Base(target) != target || filepath.Ext(target) != ".sql":
			migration.Problem = "Invalid migration filename in the selection plan."
		default:
			migration.ExitCode = result.ExitCode
			var data struct {
				RuntimeResult *models.RuntimeAnalysisResult `json:"runtimeResult"`
			}
			// Derive the path from the validated target, never from an artifact's
			// result field. The publisher only reads files in its results directory.
			err := readJSON(filepath.Join(output, target+".json"), &data)
			if err != nil || data.RuntimeResult == nil || data.RuntimeResult.MigrationID != target || data.RuntimeResult.Verdict == "" {
				migration.Problem = fmt.Sprintf("Runtime output is missing or invalid (CLI exit code %d). Check the diagnostic logs.", result.ExitCode)
			} else {
				migration.Runtime = data.RuntimeResult
			}
		}
		summary.Migrations = append(summary.Migrations, migration)
	}
	return summary
}
