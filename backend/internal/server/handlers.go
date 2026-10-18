package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"migration-lab/backend/internal/analysis/runtime"
	"migration-lab/backend/internal/loader"
	"migration-lab/backend/internal/models"
	"migration-lab/backend/internal/parser"

	"github.com/labstack/echo/v4"
)

func datasetsHandler(store MigrationStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string][]string{"datasets": store.DatasetIds()})
	}
}

func migrationsHandler(store MigrationStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		if path := c.QueryParam("migrationsPath"); path != "" {
			return migrationsFromPath(c, path)
		}

		infos, err := store.MigrationInfos(c.QueryParam("datasetId"))
		if errors.Is(err, models.ErrNotFound) {
			return notFound(c)
		}
		if err != nil {
			return internalServerError(c, err)
		}

		migrations, err := parser.ParseTimeline(infos)
		if err != nil {
			return internalServerError(c, err)
		}
		return c.JSON(http.StatusOK, toResponse(migrations))
	}
}

// migrationsFromPath loads *.sql migrations from a local directory, parses and
// analyzes them, and returns the timeline response. Used for local development.
func migrationsFromPath(c echo.Context, path string) error {
	infos, err := loader.LoadMigrationInfosFromDir(path)
	if err != nil {
		return badRequest(c, err)
	}

	migrations, err := parser.ParseTimeline(infos)
	if err != nil {
		return badRequest(c, err)
	}

	return c.JSON(http.StatusOK, toResponse(migrations))
}

func runMigrationsHandler(run RunDatasetMigrations) echo.HandlerFunc {
	return func(c echo.Context) error {
		result, err := run(c.Request().Context(), c.QueryParam("datasetId"))
		if errors.Is(err, models.ErrNotFound) {
			return notFound(c)
		}
		if err != nil {
			return internalServerError(c, err)
		}
		return c.JSON(http.StatusOK, result)
	}
}

// runtimeHandler measures the migration named by the migration parameter, or
// every migration of the timeline when it is absent, against a seeded container.
// It is a POST because it starts a container and writes to it.
func runtimeHandler(store MigrationStore, analyse RuntimeAnalyse) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := queryInt64(c, "rows")
		if err != nil {
			return badRequest(c, err)
		}
		deadlineMs, err := queryInt64(c, "deadlineMs")
		if err != nil {
			return badRequest(c, err)
		}

		migrations, err := migrationSources(store, c.QueryParam("datasetId"), c.QueryParam("migrationsPath"))
		if errors.Is(err, models.ErrNotFound) {
			return notFound(c)
		}
		if err != nil {
			return badRequest(c, err)
		}

		result, err := analyse(c.Request().Context(), runtime.Request{
			Migrations:        migrations,
			TargetMigrationID: c.QueryParam("migration"),
			Rows:              rows,
			Deadline:          time.Duration(deadlineMs) * time.Millisecond,
		})
		if errors.Is(err, models.ErrNotFound) {
			return notFound(c)
		}
		if err != nil {
			return badRequest(c, err)
		}
		return c.JSON(http.StatusOK, result)
	}
}

// migrationSources resolves the request parameters to the raw migrations to
// analyse: a local directory when migrationsPath is given, else the dataset.
func migrationSources(store MigrationStore, datasetId, migrationsPath string) ([]models.MigrationInfo, error) {
	if migrationsPath != "" {
		return loader.LoadMigrationInfosFromDir(migrationsPath)
	}
	return store.MigrationInfos(datasetId)
}

// queryInt64 reads an optional numeric query parameter. Absent means zero, which
// the analysis reads as "use the default"; present but not a number is an error,
// so a typo cannot quietly turn into a different measurement than the one asked
// for.
func queryInt64(c echo.Context, name string) (int64, error) {
	raw := c.QueryParam(name)
	if raw == "" {
		return 0, nil
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a whole number, got %q", name, raw)
	}
	return value, nil
}

func notFound(c echo.Context) error {
	return c.JSON(http.StatusNotFound, map[string]string{
		"error": "unknown datasetId",
	})
}

func badRequest(c echo.Context, err error) error {
	return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
}
