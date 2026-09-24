package runner

import (
	"context"
	"fmt"

	"migration-lab/backend/internal/models"
	"migration-lab/backend/internal/monitoring"
	"migration-lab/backend/internal/pg"

	"github.com/jackc/pgx/v5"
)

func RunMigrations(ctx context.Context, migrations []models.MigrationInfo) models.RunMigrationsResult {
	monitoring.Infof(ctx, "Starting migrations: %d total", len(migrations))

	connStr, terminate, err := pg.Start(ctx)
	if err != nil {
		return models.RunMigrationsResult{
			Success:           false,
			Message:           err.Error(),
			MigrationsApplied: 0,
		}
	}
	defer terminate()

	monitoring.Infof(ctx, "PostgreSQL container started")

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		monitoring.CaptureError(ctx, err)
		return models.RunMigrationsResult{
			Success:           false,
			Message:           fmt.Sprintf("Failed to connect to database: %v", err),
			MigrationsApplied: 0,
		}
	}
	defer func() {
		if err := conn.Close(context.WithoutCancel(ctx)); err != nil {
			monitoring.Errorf(ctx, "Failed to close migration runner connection: %v", err)
		}
	}()

	applied := 0
	for _, m := range migrations {
		monitoring.Infof(ctx, "Applying migration: %s", m.ID)

		_, err := conn.Exec(ctx, m.SQL)
		if err != nil {
			monitoring.Errorf(ctx, "Migration %s failed: %v", m.ID, err)
			return models.RunMigrationsResult{
				Success:           false,
				Message:           fmt.Sprintf("Failed at %s: %v", m.ID, err),
				MigrationsApplied: applied,
			}
		}
		applied++
		monitoring.Infof(ctx, "Migration %s applied successfully", m.ID)
	}

	monitoring.Infof(ctx, "All %d migrations applied successfully", applied)
	return models.RunMigrationsResult{
		Success:           true,
		Message:           "All migrations applied successfully",
		MigrationsApplied: applied,
	}
}
