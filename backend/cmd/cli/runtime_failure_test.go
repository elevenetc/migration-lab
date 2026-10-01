package main

import (
	"strings"
	"testing"

	"migration-lab/backend/internal/models"
)

func TestRuntimeFailure(t *testing.T) {
	cancelled := models.RuntimeMigrationResult{MigrationID: "V2__add.sql", Verdict: models.RuntimeExceedsDeadline, Message: "statement 0 was cancelled"}
	completed := models.RuntimeMigrationResult{MigrationID: "V1__create.sql", Verdict: models.RuntimeCompleted}

	for _, test := range []struct {
		name              string
		result            *models.RuntimeAnalysisResult
		targetMigrationID string
		want              string
	}{
		{"no runtime pass", nil, "", ""},
		{"run could not start", &models.RuntimeAnalysisResult{Migrations: []models.RuntimeMigrationResult{}, Message: "docker unavailable"}, "", "could not start: docker unavailable"},
		{"every migration reports without failing", &models.RuntimeAnalysisResult{Migrations: []models.RuntimeMigrationResult{completed, cancelled}}, "", ""},
		{"target missed the deadline", &models.RuntimeAnalysisResult{Migrations: []models.RuntimeMigrationResult{cancelled}}, "V2__add.sql", "V2__add.sql: EXCEEDS_DEADLINE"},
		{"target completed", &models.RuntimeAnalysisResult{Migrations: []models.RuntimeMigrationResult{completed}}, "V1__create.sql", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := runtimeFailure(test.result, test.targetMigrationID)
			if test.want == "" && err != nil {
				t.Fatalf("expected no failure, got %v", err)
			}
			if test.want != "" && (err == nil || !strings.Contains(err.Error(), test.want)) {
				t.Fatalf("expected %q, got %v", test.want, err)
			}
		})
	}
}
