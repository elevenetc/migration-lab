package parser

import (
	"testing"

	"migration-timeline/backend/internal/models"
)

func TestParseCreateTable(t *testing.T) {
	sql := `CREATE TABLE users (id SERIAL PRIMARY KEY, name VARCHAR(255) NOT NULL);`

	result, err := ParseCreateTable("test-migration", sql)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.MigrationID != "test-migration" {
		t.Errorf("expected migrationID 'test-migration', got '%s'", result.MigrationID)
	}

	if result.TableName != "users" {
		t.Errorf("expected table name 'users', got '%s'", result.TableName)
	}

	if len(result.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(result.Columns))
	}

	idCol := result.Columns[0]
	if idCol.Name != "id" {
		t.Errorf("expected first column name 'id', got '%s'", idCol.Name)
	}
	if idCol.Type != "serial" {
		t.Errorf("expected first column type 'serial', got '%s'", idCol.Type)
	}
	if !containsConstraint(idCol.Constraints, "PRIMARY KEY") {
		t.Errorf("expected PRIMARY KEY constraint on id column, got %v", idCol.Constraints)
	}

	nameCol := result.Columns[1]
	if nameCol.Name != "name" {
		t.Errorf("expected second column name 'name', got '%s'", nameCol.Name)
	}
	if nameCol.Type != "varchar(255)" {
		t.Errorf("expected second column type 'varchar(255)', got '%s'", nameCol.Type)
	}
	if !containsConstraint(nameCol.Constraints, "NOT NULL") {
		t.Errorf("expected NOT NULL constraint on name column, got %v", nameCol.Constraints)
	}
}

func TestParseMigration_CreateTable(t *testing.T) {
	m := parseMigrationTest(t, "V1__create_users.sql", `CREATE TABLE users (id SERIAL PRIMARY KEY);`, 1)
	if m.ID != "V1__create_users.sql" {
		t.Errorf("expected ID 'V1__create_users.sql', got '%s'", m.ID)
	}
	if m.Version != "1" {
		t.Errorf("expected version '1', got '%s'", m.Version)
	}
	op := m.Operations[0].(models.CreateTable)
	if op.TableName != "users" {
		t.Errorf("expected table name 'users', got '%s'", op.TableName)
	}
}

func TestParseMigration_AddColumn(t *testing.T) {
	op := parseSingleOp(t, `ALTER TABLE users ADD COLUMN email VARCHAR(255) NOT NULL;`).(models.AddColumn)
	if op.TableName != "users" {
		t.Errorf("expected table name 'users', got '%s'", op.TableName)
	}
	if op.Column.Name != "email" {
		t.Errorf("expected column name 'email', got '%s'", op.Column.Name)
	}
	if op.Column.Type != "varchar(255)" {
		t.Errorf("expected column type 'varchar(255)', got '%s'", op.Column.Type)
	}
}

func TestParseMigration_DropColumn(t *testing.T) {
	op := parseSingleOp(t, `ALTER TABLE users DROP COLUMN email;`).(models.DropColumn)
	if op.TableName != "users" {
		t.Errorf("expected table name 'users', got '%s'", op.TableName)
	}
	if op.ColumnName != "email" {
		t.Errorf("expected column name 'email', got '%s'", op.ColumnName)
	}
}

func TestParseMigration_DropTable(t *testing.T) {
	op := parseSingleOp(t, `DROP TABLE users;`).(models.DropTable)
	if op.TableName != "users" {
		t.Errorf("expected table name 'users', got '%s'", op.TableName)
	}
}

func TestParseMigration_AlterColumnType(t *testing.T) {
	op := parseSingleOp(t, `ALTER TABLE users ALTER COLUMN name TYPE TEXT;`).(models.AlterColumnType)
	if op.TableName != "users" {
		t.Errorf("expected table name 'users', got '%s'", op.TableName)
	}
	if op.ColumnName != "name" {
		t.Errorf("expected column name 'name', got '%s'", op.ColumnName)
	}
	if op.NewType != "text" {
		t.Errorf("expected new type 'text', got '%s'", op.NewType)
	}
}

func TestParseMigration_SetNotNull(t *testing.T) {
	op := parseSingleOp(t, `ALTER TABLE users ALTER COLUMN name SET NOT NULL;`).(models.SetNotNull)
	if op.TableName != "users" {
		t.Errorf("expected table name 'users', got '%s'", op.TableName)
	}
	if op.ColumnName != "name" {
		t.Errorf("expected column name 'name', got '%s'", op.ColumnName)
	}
}

func TestParseMigration_DropNotNull(t *testing.T) {
	op := parseSingleOp(t, `ALTER TABLE users ALTER COLUMN name DROP NOT NULL;`).(models.DropNotNull)
	if op.TableName != "users" {
		t.Errorf("expected table name 'users', got '%s'", op.TableName)
	}
	if op.ColumnName != "name" {
		t.Errorf("expected column name 'name', got '%s'", op.ColumnName)
	}
}

func TestParseMigration_SetDefault(t *testing.T) {
	op := parseSingleOp(t, `ALTER TABLE users ALTER COLUMN status SET DEFAULT 'active';`).(models.SetDefault)
	if op.TableName != "users" {
		t.Errorf("expected table name 'users', got '%s'", op.TableName)
	}
	if op.ColumnName != "status" {
		t.Errorf("expected column name 'status', got '%s'", op.ColumnName)
	}
	if op.DefaultValue != "'active'" {
		t.Errorf("expected default value \"'active'\", got '%s'", op.DefaultValue)
	}
}

func TestParseMigration_DropDefault(t *testing.T) {
	op := parseSingleOp(t, `ALTER TABLE users ALTER COLUMN status DROP DEFAULT;`).(models.DropDefault)
	if op.TableName != "users" {
		t.Errorf("expected table name 'users', got '%s'", op.TableName)
	}
	if op.ColumnName != "status" {
		t.Errorf("expected column name 'status', got '%s'", op.ColumnName)
	}
}

func TestParseMigration_RenameTable(t *testing.T) {
	op := parseSingleOp(t, `ALTER TABLE users RENAME TO customers;`).(models.RenameTable)
	if op.TableName != "users" {
		t.Errorf("expected table name 'users', got '%s'", op.TableName)
	}
	if op.NewTableName != "customers" {
		t.Errorf("expected new table name 'customers', got '%s'", op.NewTableName)
	}
}

func TestParseMigration_RenameColumn(t *testing.T) {
	op := parseSingleOp(t, `ALTER TABLE users RENAME COLUMN name TO full_name;`).(models.RenameColumn)
	if op.TableName != "users" {
		t.Errorf("expected table name 'users', got '%s'", op.TableName)
	}
	if op.ColumnName != "name" {
		t.Errorf("expected column name 'name', got '%s'", op.ColumnName)
	}
	if op.NewColumnName != "full_name" {
		t.Errorf("expected new column name 'full_name', got '%s'", op.NewColumnName)
	}
}

func TestParseMigration_AddConstraint(t *testing.T) {
	op := parseSingleOp(t, `ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email);`).(models.AddConstraint)
	if op.TableName != "users" {
		t.Errorf("expected table name 'users', got '%s'", op.TableName)
	}
	if op.ConstraintName != "users_email_unique" {
		t.Errorf("expected constraint name 'users_email_unique', got '%s'", op.ConstraintName)
	}
	if op.ConstraintType != "UNIQUE" {
		t.Errorf("expected constraint type 'UNIQUE', got '%s'", op.ConstraintType)
	}
}

func TestParseMigration_DropConstraint(t *testing.T) {
	op := parseSingleOp(t, `ALTER TABLE users DROP CONSTRAINT users_email_unique;`).(models.DropConstraint)
	if op.TableName != "users" {
		t.Errorf("expected table name 'users', got '%s'", op.TableName)
	}
	if op.ConstraintName != "users_email_unique" {
		t.Errorf("expected constraint name 'users_email_unique', got '%s'", op.ConstraintName)
	}
}

func TestParseMigration_MultiStatement(t *testing.T) {
	m := parseMigrationTest(t, "test", `
		CREATE TABLE users (id SERIAL PRIMARY KEY);
		ALTER TABLE users ADD COLUMN name VARCHAR(255);
		ALTER TABLE users ADD COLUMN email VARCHAR(255);
	`, 0)
	if len(m.Operations) != 3 {
		t.Fatalf("expected 3 operations, got %d", len(m.Operations))
	}
	_ = m.Operations[0].(models.CreateTable)
	_ = m.Operations[1].(models.AddColumn)
	_ = m.Operations[2].(models.AddColumn)
}

func TestParseMigration_DropTableIfExists(t *testing.T) {
	op := parseSingleOp(t, `DROP TABLE IF EXISTS users;`).(models.DropTable)
	if op.TableName != "users" {
		t.Errorf("expected table name 'users', got '%s'", op.TableName)
	}
}

func TestParseMigration_DropMultipleTables(t *testing.T) {
	m := parseMigrationTest(t, "test", `DROP TABLE users, orders;`, 0)
	if len(m.Operations) != 2 {
		t.Fatalf("expected 2 operations, got %d", len(m.Operations))
	}
	op1 := m.Operations[0].(models.DropTable)
	op2 := m.Operations[1].(models.DropTable)
	if op1.TableName != "users" {
		t.Errorf("expected table name 'users', got '%s'", op1.TableName)
	}
	if op2.TableName != "orders" {
		t.Errorf("expected table name 'orders', got '%s'", op2.TableName)
	}
}

func TestParseMigration_ForeignKeyConstraint(t *testing.T) {
	op := parseSingleOp(t, `ALTER TABLE orders ADD CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id);`).(models.AddConstraint)
	if op.ConstraintName != "fk_user" {
		t.Errorf("expected constraint name 'fk_user', got '%s'", op.ConstraintName)
	}
	if op.ConstraintType != "FOREIGN_KEY" {
		t.Errorf("expected constraint type 'FOREIGN_KEY', got '%s'", op.ConstraintType)
	}
}

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"V1__create_users.sql", "1"},
		{"V1_2__add_column.sql", "1.2"},
		{"V1_2_3__migration.sql", "1.2.3"},
		{"1__init.sql", "1"},
		{"unknown.sql", "unknown"},
	}

	for _, tt := range tests {
		result := extractVersion(tt.input)
		if result != tt.expected {
			t.Errorf("extractVersion(%s) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

func containsConstraint(constraints []string, target string) bool {
	for _, c := range constraints {
		if c == target {
			return true
		}
	}
	return false
}

func parseSingleOp(t *testing.T, sql string) models.Operation {
	t.Helper()
	return parseMigrationTest(t, "test", sql, 0).Operations[0]
}

func parseMigrationTest(t *testing.T, id, sql string, timestamp int64) *models.Migration {
	t.Helper()
	result, err := ParseMigration(id, sql, timestamp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return result
}
