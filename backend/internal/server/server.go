package server

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Config struct {
	Port                  int
	MigrationsProvider    MigrationsProvider
	MigrationsRunProvider MigrationsRunProvider
}

func New(cfg Config) *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{echo.GET, echo.POST},
	}))

	e.GET("/health", healthHandler)
	e.GET("/api/migrations", migrationsHandler(cfg.MigrationsProvider))
	if cfg.MigrationsRunProvider != nil {
		e.POST("/api/migrations/run", runMigrationsHandler(cfg.MigrationsRunProvider))
	}

	return e
}

func healthHandler(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func Start(cfg Config) error {
	e := New(cfg)
	return e.Start(fmt.Sprintf(":%d", cfg.Port))
}
