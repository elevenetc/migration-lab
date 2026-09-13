package parser

import (
	"testing"

	"migration-lab/backend/internal/models"
)

func timeline(t *testing.T, sqls ...string) []*models.Migration {
	t.Helper()
	infos := make([]models.MigrationInfo, 0, len(sqls))
	for i, sql := range sqls {
		infos = append(infos, models.MigrationInfo{ID: "v" + string(rune('1'+i)), SQL: sql, Timestamp: int64(i + 1)})
	}
	migrations, err := ParseTimeline(infos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return migrations
}

func alterColumnTypeOf(t *testing.T, migration *models.Migration) models.AlterColumnType {
	t.Helper()
	for _, statement := range migration.Statements {
		for _, operation := range statement.Operations {
			if op, ok := operation.(models.AlterColumnType); ok {
				return op
			}
		}
	}
	t.Fatalf("no ALTER COLUMN TYPE in %s", migration.ID)
	return models.AlterColumnType{}
}

func TestParseTimelineResolvesTheTypeAColumnHadBefore(t *testing.T) {
	migrations := timeline(t,
		`CREATE TABLE accounts (id INT, email VARCHAR(255));`,
		`ALTER TABLE accounts ALTER COLUMN email TYPE TEXT;`,
	)

	op := alterColumnTypeOf(t, migrations[1])
	if op.PreviousType.String() != "varchar(255)" {
		t.Errorf("expected old type 'varchar(255)', got '%s'", op.PreviousType)
	}
	if op.NewType.String() != "text" {
		t.Errorf("expected new type 'text', got '%s'", op.NewType)
	}
}

func TestParseTimelineFollowsRenamesAndAddedColumns(t *testing.T) {
	migrations := timeline(t,
		`CREATE TABLE accounts (id INT);`,
		`ALTER TABLE accounts ADD COLUMN note VARCHAR(100);`,
		`ALTER TABLE accounts RENAME COLUMN note TO comment;`,
		`ALTER TABLE accounts RENAME TO users;`,
		`ALTER TABLE users ALTER COLUMN comment TYPE TEXT;`,
	)

	if op := alterColumnTypeOf(t, migrations[4]); op.PreviousType.String() != "varchar(100)" {
		t.Errorf("expected old type 'varchar(100)', got '%s'", op.PreviousType)
	}
}

func TestParseTimelineLeavesAnUndeclaredColumnUnresolved(t *testing.T) {
	migrations := timeline(t, `ALTER TABLE accounts ALTER COLUMN email TYPE TEXT;`)

	if op := alterColumnTypeOf(t, migrations[0]); op.PreviousType.String() != "" {
		t.Errorf("expected no old type for a column no migration declared, got '%s'", op.PreviousType)
	}
}

// A second ALTER COLUMN TYPE has to see what the first one left, not what the
// table was created with.
func TestParseTimelineChainsTypeChanges(t *testing.T) {
	migrations := timeline(t,
		`CREATE TABLE accounts (email VARCHAR(50));`,
		`ALTER TABLE accounts ALTER COLUMN email TYPE VARCHAR(255);`,
		`ALTER TABLE accounts ALTER COLUMN email TYPE VARCHAR(100);`,
	)

	if op := alterColumnTypeOf(t, migrations[2]); op.PreviousType.String() != "varchar(255)" {
		t.Errorf("expected old type 'varchar(255)', got '%s'", op.PreviousType)
	}
}
