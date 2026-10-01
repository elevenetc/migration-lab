package runner

import (
	"context"
	"errors"
	"testing"

	"migration-lab/backend/internal/models"
)

func TestLoadAndRunMigrationsPropagatesNotFound(t *testing.T) {
	load := func(string) ([]models.MigrationInfo, error) { return nil, models.ErrNotFound }

	_, err := LoadAndRunMigrations(context.Background(), load, "bogus")

	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
