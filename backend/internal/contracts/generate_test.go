package contracts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"migration-timeline/backend/internal/analysis"
	"migration-timeline/backend/internal/models"
	"migration-timeline/backend/internal/parser"
)

func singleStatement(kind, sql string, ops ...models.Operation) []models.Statement {
	return []models.Statement{{Index: 0, Kind: kind, SQL: sql, Operations: ops}}
}

func TestGenerateMigrationResponseFixture(t *testing.T) {
	migrations := []*models.Migration{
		{
			ID:        "migration-1",
			Version:   "V1__create_users",
			Timestamp: 1000,
			Statements: singleStatement("CREATE_TABLE",
				"CREATE TABLE users (id SERIAL PRIMARY KEY, email VARCHAR(255) NOT NULL UNIQUE, created_at TIMESTAMP DEFAULT NOW())",
				models.CreateTable{
					TableName: "users",
					Columns: []models.Column{
						{Name: "id", Type: "SERIAL", Constraints: []string{"PRIMARY KEY"}},
						{Name: "email", Type: "VARCHAR(255)", Constraints: []string{"NOT NULL", "UNIQUE"}},
						{Name: "created_at", Type: "TIMESTAMP", Constraints: []string{"DEFAULT"}, DefaultExpr: "now()"},
					},
					IsPartitioned: false,
				}),
		},
		{
			ID:        "migration-2",
			Version:   "V2__add_user_status",
			Timestamp: 2000,
			Statements: singleStatement("ALTER_TABLE",
				"ALTER TABLE users ADD COLUMN status VARCHAR(50) DEFAULT 'active'",
				models.AddColumn{
					TableName: "users",
					Column: models.Column{
						Name: "status", Type: "VARCHAR(50)", Constraints: []string{"DEFAULT"}, DefaultExpr: "'active'",
					},
				}),
		},
		{
			ID:        "migration-3",
			Version:   "V3__change_email_type",
			Timestamp: 3000,
			Statements: singleStatement("ALTER_TABLE",
				"ALTER TABLE users ALTER COLUMN email TYPE TEXT",
				models.AlterColumnType{
					TableName:  "users",
					ColumnName: "email",
					NewType:    "TEXT",
				}),
		},
		{
			ID:        "migration-4",
			Version:   "V4__set_email_not_null",
			Timestamp: 4000,
			Statements: singleStatement("ALTER_TABLE",
				"ALTER TABLE users ALTER COLUMN email SET NOT NULL",
				models.SetNotNull{
					TableName:  "users",
					ColumnName: "email",
				}),
		},
		{
			ID:        "migration-5",
			Version:   "V5__drop_status_not_null",
			Timestamp: 5000,
			Statements: singleStatement("ALTER_TABLE",
				"ALTER TABLE users ALTER COLUMN status DROP NOT NULL",
				models.DropNotNull{
					TableName:  "users",
					ColumnName: "status",
				}),
		},
		{
			ID:        "migration-6",
			Version:   "V6__set_status_default",
			Timestamp: 6000,
			Statements: singleStatement("ALTER_TABLE",
				"ALTER TABLE users ALTER COLUMN status SET DEFAULT 'active'",
				models.SetDefault{
					TableName:    "users",
					ColumnName:   "status",
					DefaultValue: "'active'",
				}),
		},
		{
			ID:        "migration-7",
			Version:   "V7__drop_status_default",
			Timestamp: 7000,
			Statements: singleStatement("ALTER_TABLE",
				"ALTER TABLE users ALTER COLUMN status DROP DEFAULT",
				models.DropDefault{
					TableName:  "users",
					ColumnName: "status",
				}),
		},
		{
			ID:        "migration-8",
			Version:   "V8__rename_users_to_accounts",
			Timestamp: 8000,
			Statements: singleStatement("RENAME",
				"ALTER TABLE users RENAME TO accounts",
				models.RenameTable{
					TableName:    "users",
					NewTableName: "accounts",
				}),
		},
		{
			ID:        "migration-9",
			Version:   "V9__rename_email_to_email_address",
			Timestamp: 9000,
			Statements: singleStatement("RENAME",
				"ALTER TABLE accounts RENAME COLUMN email TO email_address",
				models.RenameColumn{
					TableName:     "accounts",
					ColumnName:    "email",
					NewColumnName: "email_address",
				}),
		},
		{
			ID:        "migration-10",
			Version:   "V10__add_unique_constraint",
			Timestamp: 10000,
			Statements: singleStatement("ALTER_TABLE",
				"ALTER TABLE accounts ADD CONSTRAINT uk_accounts_email UNIQUE (email_address)",
				models.AddConstraint{
					TableName:      "accounts",
					ConstraintName: "uk_accounts_email",
					ConstraintType: "UNIQUE",
				}),
		},
		{
			ID:        "migration-11",
			Version:   "V11__drop_unique_constraint",
			Timestamp: 11000,
			Statements: singleStatement("ALTER_TABLE",
				"ALTER TABLE accounts DROP CONSTRAINT uk_accounts_email",
				models.DropConstraint{
					TableName:      "accounts",
					ConstraintName: "uk_accounts_email",
				}),
		},
		{
			ID:        "migration-12",
			Version:   "V12__drop_accounts",
			Timestamp: 12000,
			Statements: singleStatement("DROP_TABLE",
				"DROP TABLE accounts",
				models.DropTable{
					TableName: "accounts",
				}),
		},
		{
			ID:        "migration-13",
			Version:   "V13__drop_legacy_column",
			Timestamp: 13000,
			Statements: singleStatement("ALTER_TABLE",
				"ALTER TABLE users DROP COLUMN legacy_field",
				models.DropColumn{
					TableName:  "users",
					ColumnName: "legacy_field",
				}),
		},
		{
			ID:        "migration-14",
			Version:   "V14__create_partitioned_table",
			Timestamp: 14000,
			Statements: singleStatement("CREATE_TABLE",
				"CREATE TABLE measurements (id SERIAL, created_at TIMESTAMP NOT NULL, value NUMERIC) PARTITION BY RANGE (created_at)",
				models.CreateTable{
					TableName: "measurements",
					Columns: []models.Column{
						{Name: "id", Type: "SERIAL", Constraints: []string{}},
						{Name: "created_at", Type: "TIMESTAMP", Constraints: []string{"NOT NULL"}},
						{Name: "value", Type: "NUMERIC", Constraints: []string{}},
					},
					IsPartitioned: true,
				}),
		},
		{
			ID:        "migration-15",
			Version:   "V15__create_partition",
			Timestamp: 15000,
			Statements: singleStatement("CREATE_TABLE",
				"CREATE TABLE measurements_2024 PARTITION OF measurements FOR VALUES FROM ('2024-01-01') TO ('2025-01-01')",
				models.CreateTable{
					TableName:     "measurements_2024",
					Columns:       []models.Column{},
					IsPartitioned: false,
					PartitionOf:   strPtr("measurements"),
				}),
		},
		{
			ID:        "migration-16",
			Version:   "V16__add_not_valid_constraint",
			Timestamp: 16000,
			Statements: singleStatement("ALTER_TABLE",
				"ALTER TABLE users ADD CONSTRAINT users_age_positive CHECK (age > 0) NOT VALID",
				models.AddConstraint{
					TableName:      "users",
					ConstraintName: "users_age_positive",
					ConstraintType: "CHECK",
					NotValid:       true,
				}),
		},
		{
			ID:        "migration-17",
			Version:   "V17__add_column_volatile_default",
			Timestamp: 17000,
			Statements: singleStatement("ALTER_TABLE",
				"ALTER TABLE users ADD COLUMN created_at TIMESTAMP DEFAULT now()",
				models.AddColumn{
					TableName: "users",
					Column: models.Column{
						Name:        "created_at",
						Type:        "TIMESTAMP",
						Constraints: []string{"DEFAULT"},
						DefaultExpr: "now()",
					},
				}),
		},
		{
			ID:        "migration-18",
			Version:   "V18__alter_partitioned",
			Timestamp: 18000,
			Statements: singleStatement("ALTER_TABLE",
				"ALTER TABLE measurements ADD COLUMN description TEXT",
				models.AddColumn{
					TableName: "measurements",
					Column:    models.Column{Name: "description", Type: "TEXT", Constraints: []string{}},
				}),
		},
	}

	// A data-only migration parses to zero supported statements. Routed through
	// ParseMigration to exercise the real serialization path: Statements must
	// marshal as [] (not null) so it stays assignable to TS `Statement[]`.
	dataOnly, err := parser.ParseMigration("V19__backfill_status",
		"UPDATE users SET status = 'active' WHERE status IS NULL;", 19000)
	if err != nil {
		t.Fatalf("Failed to parse data-only migration: %v", err)
	}
	migrations = append(migrations, dataOnly)

	createTableMap := make(map[string]models.CreateTable)
	for _, m := range migrations {
		for _, stmt := range m.Statements {
			for _, op := range stmt.Operations {
				if ct, ok := op.(models.CreateTable); ok {
					createTableMap[ct.TableName] = ct
				}
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
