package dummy

import (
	"migration-timeline/backend/internal/models"
	"migration-timeline/backend/internal/parser"
)

var simpleRenameMigrations = []struct {
	version string
	sql     string
}{
	{"v1", createUsersSimple},
	{"v2", renameUsersToClients},
}

func GetSimpleRenameMigrations() ([]*models.Migration, error) {
	var migrations []*models.Migration
	for i, m := range simpleRenameMigrations {
		migration, err := parser.ParseMigration(m.version, m.sql, int64(i+1))
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
	}
	return migrations, nil
}

func GetSimpleRenameMigrationsForRunner() []models.MigrationInfo {
	var migrations []models.MigrationInfo
	for i, m := range simpleRenameMigrations {
		migrations = append(migrations, models.MigrationInfo{
			ID:        m.version,
			SQL:       m.sql,
			Timestamp: int64(i + 1),
		})
	}
	return migrations
}

const createUsersSimple = `CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE
);`

const renameUsersToClients = `ALTER TABLE users RENAME TO clients;`
