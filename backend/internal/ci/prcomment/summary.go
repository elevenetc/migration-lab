package prcomment

import "migration-lab/backend/internal/models"

// Summary contains the collected results and diagnostics for a workflow run.
type Summary struct {
	Problem         string
	ExistingChanges int
	Migrations      []Migration
}

// Migration contains the runtime evidence or diagnostic for a selected migration.
type Migration struct {
	Name     string
	Path     string // Repository-relative Git path; empty when unavailable.
	Problem  string
	ExitCode int
	Runtime  *models.RuntimeAnalysisResult
}
