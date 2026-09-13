package database

import (
	"sort"

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

// DatasetIds returns the known dataset ids in alphabetical order.
func (d *Database) DatasetIds() []string {
	ids := make([]string, 0, len(d.datasets))
	for id := range d.datasets {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// Migrations returns the parsed migrations for the given dataset id,
// or models.ErrNotFound if the id is unknown.
func (d *Database) Migrations(migrationId string) ([]models.Migration, error) {
	infos, ok := d.datasets[migrationId]
	if !ok {
		return nil, models.ErrNotFound
	}

	parsed, err := parser.ParseTimeline(infos)
	if err != nil {
		return nil, err
	}

	migrations := make([]models.Migration, len(parsed))
	for i, m := range parsed {
		migrations[i] = *m
	}
	return migrations, nil
}
