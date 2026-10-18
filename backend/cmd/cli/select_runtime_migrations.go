package main

import (
	"fmt"

	"migration-lab/backend/internal/models"
)

// selectRuntimeMigrations retains the target and its predecessors. Later SQL
// must not affect either runtime or static analysis of the selected migration.
func selectRuntimeMigrations(infos []models.MigrationInfo, targetMigrationID string) ([]models.MigrationInfo, error) {
	for i, info := range infos {
		if info.ID == targetMigrationID {
			return infos[:i+1], nil
		}
	}
	return nil, fmt.Errorf("migration %q not found", targetMigrationID)
}
