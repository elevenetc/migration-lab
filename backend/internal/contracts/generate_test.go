package contracts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"migration-timeline/backend/internal/analysis"
	"migration-timeline/backend/internal/models"
)

func TestGenerateMigrationResponseFixture(t *testing.T) {
	migrations := []*models.Migration{
		{
			ID:        "migration-1",
			Version:   "V1__create_users",
			Timestamp: 1000,
			Operations: []models.Operation{
				models.CreateTable{
					MigrationID: "migration-1",
					TableName:   "users",
					Columns: []models.Column{
						{Name: "id", Type: "SERIAL", Constraints: []string{"PRIMARY KEY"}},
						{Name: "email", Type: "VARCHAR(255)", Constraints: []string{"NOT NULL", "UNIQUE"}},
						{Name: "created_at", Type: "TIMESTAMP", Constraints: []string{"DEFAULT NOW()"}},
					},
					IsPartitioned: false,
				},
			},
		},
		{
			ID:        "migration-2",
			Version:   "V2__add_user_status",
			Timestamp: 2000,
			Operations: []models.Operation{
				models.AddColumn{
					MigrationID: "migration-2",
					TableName:   "users",
					Column:      models.Column{Name: "status", Type: "VARCHAR(50)", Constraints: []string{"DEFAULT 'active'"}},
				},
			},
		},
		{
			ID:        "migration-3",
			Version:   "V3__change_email_type",
			Timestamp: 3000,
			Operations: []models.Operation{
				models.AlterColumnType{
					MigrationID: "migration-3",
					TableName:   "users",
					ColumnName:  "email",
					NewType:     "TEXT",
				},
			},
		},
		{
			ID:        "migration-4",
			Version:   "V4__set_email_not_null",
			Timestamp: 4000,
			Operations: []models.Operation{
				models.SetNotNull{
					MigrationID: "migration-4",
					TableName:   "users",
					ColumnName:  "email",
				},
			},
		},
		{
			ID:        "migration-5",
			Version:   "V5__drop_status_not_null",
			Timestamp: 5000,
			Operations: []models.Operation{
				models.DropNotNull{
					MigrationID: "migration-5",
					TableName:   "users",
					ColumnName:  "status",
				},
			},
		},
		{
			ID:        "migration-6",
			Version:   "V6__set_status_default",
			Timestamp: 6000,
			Operations: []models.Operation{
				models.SetDefault{
					MigrationID:  "migration-6",
					TableName:    "users",
					ColumnName:   "status",
					DefaultValue: "'active'",
				},
			},
		},
		{
			ID:        "migration-7",
			Version:   "V7__drop_status_default",
			Timestamp: 7000,
			Operations: []models.Operation{
				models.DropDefault{
					MigrationID: "migration-7",
					TableName:   "users",
					ColumnName:  "status",
				},
			},
		},
		{
			ID:        "migration-8",
			Version:   "V8__rename_users_to_accounts",
			Timestamp: 8000,
			Operations: []models.Operation{
				models.RenameTable{
					MigrationID:  "migration-8",
					TableName:    "users",
					NewTableName: "accounts",
				},
			},
		},
		{
			ID:        "migration-9",
			Version:   "V9__rename_email_to_email_address",
			Timestamp: 9000,
			Operations: []models.Operation{
				models.RenameColumn{
					MigrationID:   "migration-9",
					TableName:     "accounts",
					ColumnName:    "email",
					NewColumnName: "email_address",
				},
			},
		},
		{
			ID:        "migration-10",
			Version:   "V10__add_unique_constraint",
			Timestamp: 10000,
			Operations: []models.Operation{
				models.AddConstraint{
					MigrationID:    "migration-10",
					TableName:      "accounts",
					ConstraintName: "uk_accounts_email",
					ConstraintType: "UNIQUE",
				},
			},
		},
		{
			ID:        "migration-11",
			Version:   "V11__drop_unique_constraint",
			Timestamp: 11000,
			Operations: []models.Operation{
				models.DropConstraint{
					MigrationID:    "migration-11",
					TableName:      "accounts",
					ConstraintName: "uk_accounts_email",
				},
			},
		},
		{
			ID:        "migration-12",
			Version:   "V12__drop_accounts",
			Timestamp: 12000,
			Operations: []models.Operation{
				models.DropTable{
					MigrationID: "migration-12",
					TableName:   "accounts",
				},
			},
		},
		{
			ID:        "migration-13",
			Version:   "V13__drop_legacy_column",
			Timestamp: 13000,
			Operations: []models.Operation{
				models.DropColumn{
					MigrationID: "migration-13",
					TableName:   "users",
					ColumnName:  "legacy_field",
				},
			},
		},
		{
			ID:        "migration-14",
			Version:   "V14__create_partitioned_table",
			Timestamp: 14000,
			Operations: []models.Operation{
				models.CreateTable{
					MigrationID: "migration-14",
					TableName:   "measurements",
					Columns: []models.Column{
						{Name: "id", Type: "SERIAL", Constraints: []string{}},
						{Name: "created_at", Type: "TIMESTAMP", Constraints: []string{"NOT NULL"}},
						{Name: "value", Type: "NUMERIC", Constraints: []string{}},
					},
					IsPartitioned: true,
				},
			},
		},
		{
			ID:        "migration-15",
			Version:   "V15__create_partition",
			Timestamp: 15000,
			Operations: []models.Operation{
				models.CreateTable{
					MigrationID:   "migration-15",
					TableName:     "measurements_2024",
					Columns:       []models.Column{},
					IsPartitioned: false,
					PartitionOf:   strPtr("measurements"),
				},
			},
		},
		{
			ID:        "migration-16",
			Version:   "V16__alter_partitioned",
			Timestamp: 16000,
			Operations: []models.Operation{
				models.AddColumn{
					MigrationID: "migration-16",
					TableName:   "measurements",
					Column:      models.Column{Name: "description", Type: "TEXT", Constraints: []string{}},
				},
			},
		},
	}

	createTableMap := make(map[string]models.CreateTable)
	for _, m := range migrations {
		for _, op := range m.Operations {
			if ct, ok := op.(models.CreateTable); ok {
				createTableMap[ct.TableName] = ct
			}
		}
	}

	migrationMap := make(map[string]*models.Migration)
	for _, m := range migrations {
		migrationMap[m.ID] = m
	}

	analysisResult := analysis.Analyse(migrations)

	response := models.MigrationTimelineResponse{
		Timeline:       migrations,
		Map:            migrationMap,
		CreateTableMap: createTableMap,
		Analysis:       analysisResult,
	}

	data, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	outputPath := filepath.Join("..", "..", "..", "api-contracts", "fixtures", "migration-response.json")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		t.Fatalf("Failed to write fixture: %v", err)
	}
}

func TestGenerateRunMigrationsResultFixture(t *testing.T) {
	result := models.RunMigrationsResult{
		Success:           true,
		Message:           "All migrations applied successfully",
		MigrationsApplied: 4,
	}

	data, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	outputPath := filepath.Join("..", "..", "..", "api-contracts", "fixtures", "run-migrations-result.json")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		t.Fatalf("Failed to write fixture: %v", err)
	}
}

func strPtr(s string) *string {
	return &s
}
