package ci

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"migration-lab/backend/internal/ci/prcomment"
	"migration-lab/backend/internal/models"
)

// Missing or partial artifacts become visible diagnostics instead of leaving an
// earlier successful comment on the PR after a failed run.
func loadCommentSummary(output, head string) prcomment.Summary {
	var plan migrationPlan
	if err := readJSON(filepath.Join(output, "plan.json"), &plan); err != nil {
		return prcomment.Summary{Problem: "The migration selection plan is unavailable. Check the workflow logs."}
	}
	if plan.Head != head {
		return prcomment.Summary{Problem: "The selection plan does not match the analyzed commit."}
	}
	summary := prcomment.Summary{ExistingChanges: len(plan.ExistingChanges)}
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
		migration := prcomment.Migration{Name: target}
		if plan.Directory != "" && filepath.Base(target) == target && filepath.Ext(target) == ".sql" {
			migration.Path = path.Join(plan.Directory, target)
		}
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
			measured := targetResult(data.RuntimeResult, target)
			switch {
			case err == nil && measured != nil:
				migration.Runtime = measured
			case err == nil && data.RuntimeResult != nil && data.RuntimeResult.Message != "":
				migration.Problem = fmt.Sprintf("Runtime analysis did not measure the migration: %s (CLI exit code %d).", data.RuntimeResult.Message, result.ExitCode)
			default:
				migration.Problem = fmt.Sprintf("Runtime output is missing or invalid (CLI exit code %d). Check the diagnostic logs.", result.ExitCode)
			}
		}
		summary.Migrations = append(summary.Migrations, migration)
	}
	return summary
}

// targetResult is the measurement of target in a runtime result, nil when the
// run did not measure it.
func targetResult(result *models.RuntimeAnalysisResult, target string) *models.RuntimeMigrationResult {
	if result == nil {
		return nil
	}
	for i, migration := range result.Migrations {
		if migration.MigrationID == target && migration.Verdict != "" {
			return &result.Migrations[i]
		}
	}
	return nil
}
