package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"migration-timeline/backend/internal/models"
)

type Config struct {
	Port   int
	Store  MigrationStore
	Runner MigrationRunner
}

// MigrationStore queries parsed migrations by dataset id (satisfied by *database.Database).
type MigrationStore interface {
	Migrations(migrationId string) ([]models.Migration, error)
	DatasetIds() []string
}

// MigrationRunner runs a dataset's migrations by id (satisfied by runner.MigrationRunner).
type MigrationRunner interface {
	Run(ctx context.Context, migrationId string) (models.RunMigrationsResult, error)
}

func New(cfg Config) *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{echo.GET, echo.POST},
	}))

	e.GET("/health", healthHandler)
	e.GET("/api/datasets", datasetsHandler(cfg.Store))
	e.GET("/api/migrations", migrationsHandler(cfg.Store))
	e.POST("/api/migrations/run", runMigrationsHandler(cfg.Runner))

	return e
}

func healthHandler(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func Start(cfg Config) error {
	e := New(cfg)
	return e.Start(fmt.Sprintf(":%d", cfg.Port))
}
