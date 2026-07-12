package main

import (
	"flag"
	"log"
	"path/filepath"

	"migration-timeline/backend/internal/database"
	"migration-timeline/backend/internal/datasets"
	"migration-timeline/backend/internal/loader"
	"migration-timeline/backend/internal/runner"
	"migration-timeline/backend/internal/server"
)

func main() {
	port := flag.Int("port", 8080, "Server port")
	cfg := initConfig(port)

	log.Printf("Starting server on port %d", *port)
	if err := server.Start(cfg); err != nil {
		log.Fatal(err)
	}
}

func initConfig(port *int) server.Config {
	migrationsDir := flag.String("migrations", "", "Optional directory of .sql migrations, served under its base-name id")
	flag.Parse()

	sets := datasets.Datasets()

	if *migrationsDir != "" {
		infos, err := loader.LoadMigrationInfosFromDir(*migrationsDir)
		if err != nil {
			log.Fatal(err)
		}
		sets[filepath.Base(*migrationsDir)] = infos
	}

	db := database.New(sets)
	cfg := server.Config{
		Port:   *port,
		Store:  db,
		Runner: runner.MigrationRunner{Store: db},
	}
	return cfg
}
