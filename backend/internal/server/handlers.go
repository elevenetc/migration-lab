package server

import (
	"errors"
	"net/http"

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
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	migrations := make([]*models.Migration, len(infos))
	for i, info := range infos {
		m, err := parser.ParseMigration(info.ID, info.SQL, info.Timestamp)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		migrations[i] = m
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

func notFound(c echo.Context) error {
	return c.JSON(http.StatusNotFound, map[string]string{
		"error": "unknown migrationId",
	})
}

func pointers(migrations []models.Migration) []*models.Migration {
	ptrs := make([]*models.Migration, len(migrations))
	for i := range migrations {
		ptrs[i] = &migrations[i]
	}
	return ptrs
}
