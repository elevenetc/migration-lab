package server

import (
	"net/http"

	"migration-lab/backend/internal/monitoring"

	"github.com/labstack/echo/v4"
)

func internalServerError(c echo.Context, err error) error {
	monitoring.CaptureError(c.Request().Context(), err)
	return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
}
