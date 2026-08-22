package runner

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"migration-timeline/backend/internal/models"
	"migration-timeline/backend/internal/pg"
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

	log.Printf("PostgreSQL container started: %s", connStr)

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return models.RunMigrationsResult{
			Success:           false,
			Message:           fmt.Sprintf("Failed to connect to database: %v", err),
			MigrationsApplied: 0,
		}
	}
	defer conn.Close(ctx)

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
