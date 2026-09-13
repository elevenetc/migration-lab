package parser

import (
	"encoding/json"
	"strings"
	"testing"

	"migration-lab/backend/internal/models"
)

func TestParseCreateTable(t *testing.T) {
	sql := `CREATE TABLE users (id SERIAL PRIMARY KEY, name VARCHAR(255) NOT NULL);`

	result, err := ParseCreateTable(sql)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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
	if idCol.Type.String() != "serial" {
		t.Errorf("expected first column type 'serial', got '%s'", idCol.Type)
	}
	if !containsConstraint(idCol.Constraints, "PRIMARY KEY") {
		t.Errorf("expected PRIMARY KEY constraint on id column, got %v", idCol.Constraints)
	}

	nameCol := result.Columns[1]
	if nameCol.Name != "name" {
		t.Errorf("expected second column name 'name', got '%s'", nameCol.Name)
	}
	if nameCol.Type.String() != "varchar(255)" {
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
	stmt := m.Statements[0]
	if stmt.Index != 0 {
		t.Errorf("expected statement index 0, got %d", stmt.Index)
	}
	if stmt.Kind != "CREATE_TABLE" {
		t.Errorf("expected statement kind 'CREATE_TABLE', got '%s'", stmt.Kind)
	}
	if stmt.SQL != "CREATE TABLE users (id SERIAL PRIMARY KEY)" {
		t.Errorf("unexpected statement SQL: '%s'", stmt.SQL)
	}
	op := stmt.Operations[0].(models.CreateTable)
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
	if op.Column.Type.String() != "varchar(255)" {
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
	if op.NewType.String() != "text" {
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
	if op.NotValid {
		t.Error("expected NotValid false for a validating constraint")
	}
}

func TestParseMigration_AddConstraintNotValid(t *testing.T) {
	op := parseSingleOp(t, `ALTER TABLE users ADD CONSTRAINT users_age_check CHECK (age > 0) NOT VALID;`).(models.AddConstraint)
	if op.ConstraintType != "CHECK" {
		t.Errorf("expected constraint type 'CHECK', got '%s'", op.ConstraintType)
	}
	if !op.NotValid {
		t.Error("expected NotValid true for a NOT VALID constraint")
	}
}

func TestParseMigration_AddColumnDefaultExpr(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected string
	}{
		{"no default", `ALTER TABLE users ADD COLUMN note TEXT;`, ""},
		{"string constant", `ALTER TABLE users ADD COLUMN status TEXT DEFAULT 'active';`, "'active'"},
		{"integer constant", `ALTER TABLE users ADD COLUMN retries INT DEFAULT 3;`, "3"},
		{"function call", `ALTER TABLE users ADD COLUMN created_at TIMESTAMP DEFAULT now();`, "now()"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := parseSingleOp(t, test.sql).(models.AddColumn)
			if op.Column.DefaultExpr != test.expected {
				t.Errorf("expected default expression '%s', got '%s'", test.expected, op.Column.DefaultExpr)
			}
		})
	}
}

// Volatility comes from the expression tree, not from the deparsed string: a cast
// hides the call from any suffix match, and a volatile default is the difference
// between a catalog write and a full rewrite.
func TestParseMigration_AddColumnDefaultVolatility(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		volatile bool
	}{
		{"no default", `ALTER TABLE users ADD COLUMN note TEXT;`, false},
		{"string constant", `ALTER TABLE users ADD COLUMN status TEXT DEFAULT 'active';`, false},
		{"cast constant", `ALTER TABLE users ADD COLUMN retries INT DEFAULT '3'::int;`, false},
		{"stable function", `ALTER TABLE users ADD COLUMN created_at TIMESTAMP DEFAULT now();`, false},
		{"keyword function", `ALTER TABLE users ADD COLUMN seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;`, false},
		{"stable function of a constant", `ALTER TABLE users ADD COLUMN tag TEXT DEFAULT upper('a');`, false},
		{"volatile function", `ALTER TABLE users ADD COLUMN key UUID DEFAULT gen_random_uuid();`, true},
		{"volatile behind a cast", `ALTER TABLE users ADD COLUMN token INT DEFAULT random()::int;`, true},
		{"volatile in an expression", `ALTER TABLE users ADD COLUMN score INT DEFAULT random() * 100;`, true},
		{"volatile nested in a stable call", `ALTER TABLE users ADD COLUMN hash TEXT DEFAULT md5(random()::text);`, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := parseSingleOp(t, test.sql).(models.AddColumn)
			if op.Column.DefaultVolatile != test.volatile {
				t.Errorf("expected volatile %v, got %v", test.volatile, op.Column.DefaultVolatile)
			}
		})
	}
}

func TestParseCreateTable_ColumnDefaultExpr(t *testing.T) {
	result, err := ParseCreateTable(`CREATE TABLE users (id INT, created_at TIMESTAMP DEFAULT now());`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Columns[0].DefaultExpr != "" {
		t.Errorf("expected no default expression on id, got '%s'", result.Columns[0].DefaultExpr)
	}
	if result.Columns[1].DefaultExpr != "now()" {
		t.Errorf("expected default expression 'now()', got '%s'", result.Columns[1].DefaultExpr)
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
	if len(m.Statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(m.Statements))
	}
	for i, stmt := range m.Statements {
		if stmt.Index != i {
			t.Errorf("expected statement index %d, got %d", i, stmt.Index)
		}
		if len(stmt.Operations) != 1 {
			t.Fatalf("expected 1 operation in statement %d, got %d", i, len(stmt.Operations))
		}
	}
	if m.Statements[0].Kind != "CREATE_TABLE" {
		t.Errorf("expected statement kind 'CREATE_TABLE', got '%s'", m.Statements[0].Kind)
	}
	if m.Statements[1].Kind != "ALTER_TABLE" {
		t.Errorf("expected statement kind 'ALTER_TABLE', got '%s'", m.Statements[1].Kind)
	}
	if m.Statements[2].SQL != "ALTER TABLE users ADD COLUMN email VARCHAR(255)" {
		t.Errorf("unexpected statement SQL: '%s'", m.Statements[2].SQL)
	}
	_ = m.Statements[0].Operations[0].(models.CreateTable)
	_ = m.Statements[1].Operations[0].(models.AddColumn)
	_ = m.Statements[2].Operations[0].(models.AddColumn)
}

func TestParseMigration_DropTableIfExists(t *testing.T) {
	op := parseSingleOp(t, `DROP TABLE IF EXISTS users;`).(models.DropTable)
	if op.TableName != "users" {
		t.Errorf("expected table name 'users', got '%s'", op.TableName)
	}
}

func TestParseMigration_DropMultipleTables(t *testing.T) {
	m := parseMigrationTest(t, "test", `DROP TABLE users, orders;`, 0)
	if len(m.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(m.Statements))
	}
	stmt := m.Statements[0]
	if stmt.Kind != "DROP_TABLE" {
		t.Errorf("expected statement kind 'DROP_TABLE', got '%s'", stmt.Kind)
	}
	if len(stmt.Operations) != 2 {
		t.Fatalf("expected 2 operations, got %d", len(stmt.Operations))
	}
	op1 := stmt.Operations[0].(models.DropTable)
	op2 := stmt.Operations[1].(models.DropTable)
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

// A data-only migration parses to zero supported statements. It must still
// marshal as [] rather than null, or the response stops being assignable to the
// frontend's Statement[] — which only fails much later, in tsc.
func TestParseMigrationMarshalsNoStatementsAsAnEmptyArray(t *testing.T) {
	migration := parseMigrationTest(t, "V1__backfill", "UPDATE users SET status = 'active';", 1)

	if migration.Statements == nil {
		t.Fatal("expected an empty slice of statements, got nil")
	}

	data, err := json.Marshal(migration)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(data), `"statements":[]`) {
		t.Errorf("expected statements to marshal as [], got %s", data)
	}
}

func parseSingleOp(t *testing.T, sql string) models.Operation {
	t.Helper()
	return parseMigrationTest(t, "test", sql, 0).Statements[0].Operations[0]
}

func parseMigrationTest(t *testing.T, id, sql string, timestamp int64) *models.Migration {
	t.Helper()
	result, err := ParseMigration(id, sql, timestamp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return result
}
