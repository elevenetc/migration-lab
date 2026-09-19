package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"migration-lab/backend/internal/analysis/runtime"
	"migration-lab/backend/internal/database"
	"migration-lab/backend/internal/datasets"
	"migration-lab/backend/internal/loader"
	"migration-lab/backend/internal/monitoring"
	"migration-lab/backend/internal/runner"
	"migration-lab/backend/internal/server"

	"github.com/getsentry/sentry-go"
)

func main() {
	err := monitoring.RunWithSentry(
		os.Getenv("SENTRY_DSN"),
		os.Getenv("SENTRY_ENVIRONMENT"),
		os.Getenv("SENTRY_RELEASE"),
		run,
	)
	if err != nil {
		log.Fatal(err)
	}
}

func run(hub *sentry.Hub) error {
	port := flag.Int("port", 8080, "Server port")
	cfg, err := initConfig(port)
	if err != nil {
		return err
	}
	cfg.Sentry = hub

	log.Printf("Starting server on port %d", *port)
	return server.Start(cfg)
}

func initConfig(port *int) (server.Config, error) {
	migrationsDir := flag.String("migrations", "", "Optional directory of .sql migrations, served under its base-name id")
	flag.Parse()

	sets := datasets.Datasets()

	if *migrationsDir != "" {
		infos, err := loader.LoadMigrationInfosFromDir(*migrationsDir)
		if err != nil {
			return server.Config{}, err
		}
		sets[filepath.Base(*migrationsDir)] = infos
	}

	db := database.New(sets)
	cfg := server.Config{
		Port:    *port,
		Store:   db,
		Runner:  runner.MigrationRunner{Store: db},
		Analyse: runtime.Analyse,
	}
	return cfg, nil
}
