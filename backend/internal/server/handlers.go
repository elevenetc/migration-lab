package server

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"migration-timeline/backend/internal/models"
)

func migrationsHandler(store MigrationStore) echo.HandlerFunc {
	return func(c echo.Context) error {
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
