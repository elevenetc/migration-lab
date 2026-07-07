package runner

import (
	"context"
	"strings"
	"testing"

	"migration-timeline/backend/internal/models"
)

func TestRunMigrationsExecutesSQLSuccessfully(t *testing.T) {
	ctx := context.Background()

	migrations := []models.MigrationInfo{
		{
			ID: "V1__create_users",
			SQL: `CREATE TABLE users (
				id SERIAL PRIMARY KEY,
				name VARCHAR(255) NOT NULL
			);`,
		},
	}

	result := RunMigrations(ctx, migrations)

	if !result.Success {
		t.Errorf("Expected success but got: %s", result.Message)
	}
	if result.MigrationsApplied != 1 {
		t.Errorf("Expected 1 migration applied, got %d", result.MigrationsApplied)
	}
	if result.Message != "All migrations applied successfully" {
		t.Errorf("Unexpected message: %s", result.Message)
	}
}

func TestRunMigrationsReportsFailureOnInvalidSQL(t *testing.T) {
	ctx := context.Background()

	migrations := []models.MigrationInfo{
		{
			ID:  "V1__invalid",
			SQL: "INVALID SQL STATEMENT",
		},
	}

	result := RunMigrations(ctx, migrations)

	if result.Success {
		t.Error("Expected failure but got success")
	}
	if result.MigrationsApplied != 0 {
		t.Errorf("Expected 0 migrations applied, got %d", result.MigrationsApplied)
	}
	if !strings.Contains(result.Message, "Failed at V1__invalid") {
		t.Errorf("Expected message to contain 'Failed at V1__invalid', got: %s", result.Message)
	}
}
