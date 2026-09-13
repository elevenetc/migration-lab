package runner

import (
	"context"

	"migration-lab/backend/internal/models"
)

// Store is the migration source the runner queries (satisfied by *database.Database).
type Store interface {
	Migrations(migrationId string) ([]models.Migration, error)
}

// MigrationRunner runs a dataset's migrations against a fresh PostgreSQL container.
type MigrationRunner struct {
	Store Store
}

// Run queries the store for the dataset, then executes its SQL. It propagates
// the store's error (e.g. models.ErrNotFound) without spinning up a container.
func (r MigrationRunner) Run(ctx context.Context, migrationId string) (models.RunMigrationsResult, error) {
	migrations, err := r.Store.Migrations(migrationId)
	if err != nil {
		return models.RunMigrationsResult{}, err
	}

	infos := make([]models.MigrationInfo, len(migrations))
	for i, m := range migrations {
		infos[i] = models.MigrationInfo{ID: m.ID, SQL: m.SQL, Timestamp: m.Timestamp}
	}
	return RunMigrations(ctx, infos), nil
}
