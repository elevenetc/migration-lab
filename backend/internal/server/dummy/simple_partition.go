package dummy

import (
	"migration-timeline/backend/internal/models"
	"migration-timeline/backend/internal/parser"
)

var simplePartitionMigrations = []struct {
	version string
	sql     string
}{
	{"v1", createEventsPartitioned},
	{"v2", createEvents2024},
	{"v3", createEvents2025},
}

func GetSimplePartitionMigrations() ([]*models.Migration, error) {
	var migrations []*models.Migration
	for i, m := range simplePartitionMigrations {
		migration, err := parser.ParseMigration(m.version, m.sql, int64(i+1))
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
	}
	return migrations, nil
}

func GetSimplePartitionMigrationsForRunner() []models.MigrationInfo {
	var migrations []models.MigrationInfo
	for i, m := range simplePartitionMigrations {
		migrations = append(migrations, models.MigrationInfo{
			ID:        m.version,
			SQL:       m.sql,
			Timestamp: int64(i + 1),
		})
	}
	return migrations
}

const createEventsPartitioned = `CREATE TABLE events (
    id SERIAL,
    created_at TIMESTAMP NOT NULL
) PARTITION BY RANGE (created_at);`

const createEvents2024 = `CREATE TABLE events_2024 PARTITION OF events
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');`

const createEvents2025 = `CREATE TABLE events_2025 PARTITION OF events
    FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');`
