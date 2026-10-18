package database

import (
	"slices"
	"sort"

	"migration-lab/backend/internal/models"
)

// Database is a temporary, in-memory stand-in for a real migrations database.
// It stores raw migration sources per dataset id.
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

// MigrationInfos returns the raw migration sources of the given dataset id,
// or models.ErrNotFound if the id is unknown.
func (d *Database) MigrationInfos(datasetId string) ([]models.MigrationInfo, error) {
	infos, ok := d.datasets[datasetId]
	if !ok {
		return nil, models.ErrNotFound
	}
	return slices.Clone(infos), nil
}
