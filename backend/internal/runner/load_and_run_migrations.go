package runner

import (
	"context"

	"migration-lab/backend/internal/models"
)

// LoadMigrationInfos returns a dataset's raw migration sources by id (satisfied by
// (*database.Database).MigrationInfos).
type LoadMigrationInfos func(datasetId string) ([]models.MigrationInfo, error)

// LoadAndRunMigrations loads the migrations, then executes their SQL against a
// fresh PostgreSQL container. It propagates the load error (e.g. models.ErrNotFound)
// without spinning up a container.
func LoadAndRunMigrations(ctx context.Context, load LoadMigrationInfos, datasetId string) (models.RunMigrationsResult, error) {
	infos, err := load(datasetId)
	if err != nil {
		return models.RunMigrationsResult{}, err
	}
	return RunMigrations(ctx, infos), nil
}
