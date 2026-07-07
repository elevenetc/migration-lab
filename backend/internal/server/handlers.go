package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"migration-timeline/backend/internal/models"
	"migration-timeline/backend/internal/runner"
)

type MigrationsProvider func() ([]*models.Migration, error)
type MigrationsRunProvider func() []models.MigrationInfo

func migrationsHandler(provider MigrationsProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		migrations, err := provider()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
		response := toResponse(migrations)
		return c.JSON(http.StatusOK, response)
	}
}

func runMigrationsHandler(provider MigrationsRunProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		migrations := provider()
		result := runner.RunMigrations(c.Request().Context(), migrations)
		return c.JSON(http.StatusOK, result)
	}
}
