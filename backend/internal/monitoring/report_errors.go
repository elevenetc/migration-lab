package monitoring

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/getsentry/sentry-go"
	"github.com/getsentry/sentry-go/attribute"
	"github.com/labstack/echo/v4"
)

// ReportErrors attaches request context to Sentry events and recovers handler panics.
func ReportErrors(baseHub *sentry.Hub) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			var hub *sentry.Hub
			if baseHub != nil {
				hub = baseHub.Clone()
				hub.Scope().SetRequest(c.Request())
				for key, value := range map[string]string{
					"request_id":   c.Response().Header().Get(echo.HeaderXRequestID),
					"dataset_id":   c.QueryParam("migrationId"),
					"migration_id": c.QueryParam("migration"),
					"route":        c.Request().Method + " " + c.Path(),
				} {
					hub.Scope().SetTag(key, value)
					hub.Scope().SetAttributes(attribute.String(key, value))
				}
				c.SetRequest(c.Request().WithContext(sentry.SetHubOnContext(c.Request().Context(), hub)))
			}
			defer func() {
				if value := recover(); value != nil {
					if hub != nil {
						hub.Recover(value)
					}
					Errorf(c.Request().Context(), "panic: %v\n%s", value, debug.Stack())
					err = echo.NewHTTPError(http.StatusInternalServerError).SetInternal(fmt.Errorf("panic: %v", value))
				}
			}()
			err = next(c)
			var httpError *echo.HTTPError
			if err != nil && (!errors.As(err, &httpError) || httpError.Code >= http.StatusInternalServerError) {
				CaptureError(c.Request().Context(), err)
			}
			return err
		}
	}
}
