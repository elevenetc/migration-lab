package runner

import (
	"context"
	"fmt"
	"log"

	"migration-lab/backend/internal/models"
	"migration-lab/backend/internal/monitoring"
	"migration-lab/backend/internal/pg"

	"github.com/jackc/pgx/v5"
)

func RunMigrations(ctx context.Context, migrations []models.MigrationInfo) models.RunMigrationsResult {
	log.Printf("Starting migrations: %d total", len(migrations))

	connStr, terminate, err := pg.Start(ctx)
	if err != nil {
		return models.RunMigrationsResult{
			Success:           false,
			Message:           err.Error(),
			MigrationsApplied: 0,
		}
	}
	defer terminate()

	log.Printf("PostgreSQL container started")

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
			log.Printf("Failed to close migration runner connection: %v", err)
		}
	}()

	applied := 0
	for _, m := range migrations {
		log.Printf("Applying migration: %s", m.ID)

		_, err := conn.Exec(ctx, m.SQL)
		if err != nil {
			log.Printf("Migration %s failed: %v", m.ID, err)
			return models.RunMigrationsResult{
				Success:           false,
				Message:           fmt.Sprintf("Failed at %s: %v", m.ID, err),
				MigrationsApplied: applied,
			}
		}
		applied++
		log.Printf("Migration %s applied successfully", m.ID)
	}

	log.Printf("All %d migrations applied successfully", applied)
	return models.RunMigrationsResult{
		Success:           true,
		Message:           "All migrations applied successfully",
		MigrationsApplied: applied,
	}
}
