package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"migration-timeline/backend/internal/analysis/runtime"
	"migration-timeline/backend/internal/loader"
	"migration-timeline/backend/internal/models"
	"migration-timeline/backend/internal/parser"

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

		migrations, err := store.Migrations(c.QueryParam("migrationId"))
		if errors.Is(err, models.ErrNotFound) {
			return notFound(c)
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusOK, toResponse(pointers(migrations)))
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

func runMigrationsHandler(runner MigrationRunner) echo.HandlerFunc {
	return func(c echo.Context) error {
		result, err := runner.Run(c.Request().Context(), c.QueryParam("migrationId"))
		if errors.Is(err, models.ErrNotFound) {
			return notFound(c)
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusOK, result)
	}
}

// runtimeHandler measures one migration of the timeline against a seeded
// container. It is a POST because it starts a container and writes to it.
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

		migrations, err := migrationSources(store, c.QueryParam("migrationId"), c.QueryParam("migrationsPath"))
		if errors.Is(err, models.ErrNotFound) {
			return notFound(c)
		}
		if err != nil {
			return badRequest(c, err)
		}

		result, err := analyse(c.Request().Context(), runtime.Request{
			Migrations: migrations,
			Target:     c.QueryParam("migration"),
			Rows:       rows,
			Deadline:   time.Duration(deadlineMs) * time.Millisecond,
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
func migrationSources(store MigrationStore, migrationId, migrationsPath string) ([]models.MigrationInfo, error) {
	if migrationsPath != "" {
		return loader.LoadMigrationInfosFromDir(migrationsPath)
	}

	migrations, err := store.Migrations(migrationId)
	if err != nil {
		return nil, err
	}

	infos := make([]models.MigrationInfo, len(migrations))
	for i, migration := range migrations {
		infos[i] = models.MigrationInfo{ID: migration.ID, SQL: migration.SQL, Timestamp: migration.Timestamp}
	}
	return infos, nil
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
		"error": "unknown migrationId",
	})
}

func badRequest(c echo.Context, err error) error {
	return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
}

func pointers(migrations []models.Migration) []*models.Migration {
	ptrs := make([]*models.Migration, len(migrations))
	for i := range migrations {
		ptrs[i] = &migrations[i]
	}
	return ptrs
}
