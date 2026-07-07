package runner

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"migration-timeline/backend/internal/models"
)

func RunMigrations(ctx context.Context, migrations []models.MigrationInfo) models.RunMigrationsResult {
	log.Printf("Starting migrations: %d total", len(migrations))

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return models.RunMigrationsResult{
			Success:           false,
			Message:           fmt.Sprintf("Failed to start PostgreSQL container: %v", err),
			MigrationsApplied: 0,
		}
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			log.Printf("Failed to terminate container: %v", err)
		}
	}()

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return models.RunMigrationsResult{
			Success:           false,
			Message:           fmt.Sprintf("Failed to get connection string: %v", err),
			MigrationsApplied: 0,
		}
	}

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
