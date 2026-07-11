package runner

import (
	"context"
	"errors"
	"testing"

	"migration-timeline/backend/internal/models"
)

type stubStore struct {
	migrations []models.Migration
	err        error
}

func (s stubStore) Migrations(string) ([]models.Migration, error) {
	return s.migrations, s.err
}

func TestMigrationRunnerPropagatesNotFound(t *testing.T) {
	r := MigrationRunner{Store: stubStore{err: models.ErrNotFound}}

	_, err := r.Run(context.Background(), "bogus")

	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
