package static

import (
	"testing"

	"migration-timeline/backend/internal/models"
	"migration-timeline/backend/internal/parser"
)

func TestAlterOnPartitionedTableProducesWarning(t *testing.T) {
	createPartitioned := `
		CREATE TABLE events (
			id SERIAL,
			name VARCHAR(255),
			year INT NOT NULL
		) PARTITION BY LIST (year);
	`
	addColumn := `ALTER TABLE events ADD COLUMN description TEXT;`

	m1, _ := parser.ParseMigration("V1__create_partitioned", createPartitioned, 1)
	m2, _ := parser.ParseMigration("V2__add_column", addColumn, 2)

	result := Analyse([]*models.Migration{m1, m2})

	if len(result.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(result.Warnings))
	}

	warning, ok := result.Warnings[0].(models.AccessExclusiveLock)
	if !ok {
		t.Fatal("expected AccessExclusiveLock warning")
	}

	if warning.TableName != "events" {
		t.Errorf("expected tableName 'events', got '%s'", warning.TableName)
	}

	if warning.OperationID.MigrationID != "V2__add_column" {
		t.Errorf("expected migrationId 'V2__add_column', got '%s'", warning.OperationID.MigrationID)
	}

	if warning.OperationID.StatementIndex != 0 {
		t.Errorf("expected statementIndex 0, got %d", warning.OperationID.StatementIndex)
	}

	if warning.OperationID.OpIndex != models.StatementScoped {
		t.Errorf("expected statement-scoped opIndex -1, got %d", warning.OperationID.OpIndex)
	}

	if warning.Message == "" || len(warning.Message) < 10 {
		t.Error("expected non-empty message containing 'ACCESS EXCLUSIVE'")
	}
}

func TestAlterOnNonPartitionedTableProducesNoWarning(t *testing.T) {
	createRegular := `
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255)
		);
	`
	addColumn := `ALTER TABLE users ADD COLUMN email VARCHAR(255);`

	m1, _ := parser.ParseMigration("V1__create_table", createRegular, 1)
	m2, _ := parser.ParseMigration("V2__add_column", addColumn, 2)

	result := Analyse([]*models.Migration{m1, m2})

	if len(result.Warnings) != 0 {
		t.Fatalf("expected 0 warnings, got %d", len(result.Warnings))
	}
}

func TestMultipleAltersOnSamePartitionedTableProduceMultipleWarnings(t *testing.T) {
	createPartitioned := `
		CREATE TABLE events (
			id SERIAL,
			name VARCHAR(255),
			year INT NOT NULL
		) PARTITION BY LIST (year);
	`
	addColumn := `ALTER TABLE events ADD COLUMN description TEXT;`
	dropColumn := `ALTER TABLE events DROP COLUMN name;`

	m1, _ := parser.ParseMigration("V1__create_partitioned", createPartitioned, 1)
	m2, _ := parser.ParseMigration("V2__add_column", addColumn, 2)
	m3, _ := parser.ParseMigration("V3__drop_column", dropColumn, 3)

	result := Analyse([]*models.Migration{m1, m2, m3})

	if len(result.Warnings) != 2 {
		t.Fatalf("expected 2 warnings, got %d", len(result.Warnings))
	}

	for _, w := range result.Warnings {
		warning, ok := w.(models.AccessExclusiveLock)
		if !ok {
			t.Fatal("expected AccessExclusiveLock warning")
		}
		if warning.TableName != "events" {
			t.Errorf("expected tableName 'events', got '%s'", warning.TableName)
		}
	}
}

// The lock is acquired per statement, so several ALTER commands in one
// statement must produce a single statement-scoped warning.
func TestMultipleAlterCommandsInOneStatementProduceSingleWarning(t *testing.T) {
	createPartitioned := `
		CREATE TABLE events (
			id SERIAL,
			name VARCHAR(255),
			year INT NOT NULL
		) PARTITION BY LIST (year);
	`
	multiCommandAlter := `ALTER TABLE events ADD COLUMN description TEXT, ALTER COLUMN name SET NOT NULL;`

	m1, _ := parser.ParseMigration("V1__create_partitioned", createPartitioned, 1)
	m2, _ := parser.ParseMigration("V2__multi_alter", multiCommandAlter, 2)

	result := Analyse([]*models.Migration{m1, m2})

	if len(result.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(result.Warnings))
	}

	warning := result.Warnings[0].(models.AccessExclusiveLock)
	if warning.OperationID.OpIndex != models.StatementScoped {
		t.Errorf("expected statement-scoped opIndex -1, got %d", warning.OperationID.OpIndex)
	}
}

func TestAlterOnPartitionChildDoesNotProduceWarning(t *testing.T) {
	createPartitioned := `
		CREATE TABLE events (
			id SERIAL,
			name VARCHAR(255),
			year INT NOT NULL
		) PARTITION BY LIST (year);
	`
	createPartition := `
		CREATE TABLE events_2025 PARTITION OF events
			FOR VALUES IN (2025);
	`
	alterPartitionChild := `ALTER TABLE events_2025 ADD COLUMN local_field TEXT;`

	m1, _ := parser.ParseMigration("V1__create_partitioned", createPartitioned, 1)
	m2, _ := parser.ParseMigration("V2__create_partition", createPartition, 2)
	m3, _ := parser.ParseMigration("V3__alter_partition", alterPartitionChild, 3)

	result := Analyse([]*models.Migration{m1, m2, m3})

	if len(result.Warnings) != 0 {
		t.Fatalf("expected 0 warnings, got %d", len(result.Warnings))
	}
}
