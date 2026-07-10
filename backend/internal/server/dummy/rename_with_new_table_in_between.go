package dummy

import (
	"migration-timeline/backend/internal/models"
	"migration-timeline/backend/internal/parser"
)

var renameWithNewTableInBetweenMigrations = []struct {
	version string
	sql     string
}{
	{"v1", createUsersRWNT},
	{"v2", createSettingsRWNT},
	{"v3", renameUsersToClientsRWNT},
}

func GetRenameWithNewTableInBetweenMigrations() ([]*models.Migration, error) {
	var migrations []*models.Migration
	for i, m := range renameWithNewTableInBetweenMigrations {
		migration, err := parser.ParseMigration(m.version, m.sql, int64(i+1))
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
	}
	return migrations, nil
}

func GetRenameWithNewTableInBetweenMigrationsForRunner() []models.MigrationInfo {
	var migrations []models.MigrationInfo
	for i, m := range renameWithNewTableInBetweenMigrations {
		migrations = append(migrations, models.MigrationInfo{
			ID:        m.version,
			SQL:       m.sql,
			Timestamp: int64(i + 1),
		})
	}
	return migrations
}

const createUsersRWNT = `CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE
);`

const createSettingsRWNT = `CREATE TABLE settings (
    id SERIAL PRIMARY KEY,
    key VARCHAR(255) NOT NULL,
    value TEXT
);`

const renameUsersToClientsRWNT = `ALTER TABLE users RENAME TO clients;`
