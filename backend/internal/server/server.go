package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"migration-lab/backend/internal/analysis/runtime"
	"migration-lab/backend/internal/models"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Config struct {
	Port    int
	Store   MigrationStore
	Runner  MigrationRunner
	Analyse RuntimeAnalyse
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

// RuntimeAnalyse measures one migration of a timeline against a seeded container
// (satisfied by runtime.Analyse). A function rather than an interface: there is
// one operation, and the tests stub it without a type of their own.
type RuntimeAnalyse func(ctx context.Context, request runtime.Request) (models.RuntimeAnalysisResult, error)

func New(cfg Config) *echo.Echo {
	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{echo.GET, echo.POST},
	}))

	e.GET("/health", healthHandler)
	e.GET("/api/datasets", datasetsHandler(cfg.Store))
	e.GET("/api/migrations", migrationsHandler(cfg.Store))
	e.POST("/api/migrations/run", runMigrationsHandler(cfg.Runner))
	e.POST("/api/migrations/runtime-analysis", runtimeHandler(cfg.Store, cfg.Analyse))

	return e
}

func healthHandler(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func Start(cfg Config) error {
	e := New(cfg)
	err := e.Start(fmt.Sprintf(":%d", cfg.Port))
	return errors.Join(err, e.Close())
}
