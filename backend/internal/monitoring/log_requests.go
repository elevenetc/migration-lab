package monitoring

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// LogRequests uses route templates so query strings and local paths stay out of logs.
func LogRequests() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		HandleError:  true,
		LogMethod:    true,
		LogRoutePath: true,
		LogStatus:    true,
		LogLatency:   true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			write := Infof
			if values.Status >= http.StatusInternalServerError {
				write = Errorf
			}
			write(c.Request().Context(), "HTTP %s %s status=%d duration_ms=%d",
				values.Method, values.RoutePath, values.Status, values.Latency.Milliseconds())
			return nil
		},
	})
}
