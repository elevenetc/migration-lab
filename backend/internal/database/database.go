package database

import (
	"migration-timeline/backend/internal/models"
	"migration-timeline/backend/internal/parser"
)

// Database is a temporary, in-memory stand-in for a real migrations database.
// It stores raw migration sources per dataset id and parses them on query.
type Database struct {
	datasets map[string][]models.MigrationInfo
}

func New(datasets map[string][]models.MigrationInfo) *Database {
	return &Database{datasets: datasets}
}

// Migrations returns the parsed migrations for the given dataset id,
// or models.ErrNotFound if the id is unknown.
func (d *Database) Migrations(migrationId string) ([]models.Migration, error) {
	infos, ok := d.datasets[migrationId]
	if !ok {
		return nil, models.ErrNotFound
	}

	migrations := make([]models.Migration, len(infos))
	for i, info := range infos {
		m, err := parser.ParseMigration(info.ID, info.SQL, info.Timestamp)
		if err != nil {
			return nil, err
		}
		migrations[i] = *m
	}
	return migrations, nil
}
