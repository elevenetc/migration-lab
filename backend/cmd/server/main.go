package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"migration-timeline/backend/internal/models"
	"migration-timeline/backend/internal/parser"
	"migration-timeline/backend/internal/server"
	"migration-timeline/backend/internal/server/dummy"
)

func main() {
	port := flag.Int("port", 8080, "Server port")
	migrationsSource := flag.String("migrations", "dummy", "Migrations source (dummy)")
	flag.Parse()

	var provider server.MigrationsProvider
	var runProvider server.MigrationsRunProvider

	switch *migrationsSource {
	case "dummy":
		provider = dummy.GetComplexEcommerceMigrations
		runProvider = dummy.GetComplexEcommerceMigrationsForRunner
	default:
		// Treat as directory path
		dir := *migrationsSource
		provider = func() ([]*models.Migration, error) {
			return loadMigrationsFromDir(dir)
		}
		runProvider = nil
	}

	cfg := server.Config{
		Port:                  *port,
		MigrationsProvider:    provider,
		MigrationsRunProvider: runProvider,
	}

	log.Printf("Starting server on port %d with migrations source: %s", *port, *migrationsSource)
	if err := server.Start(cfg); err != nil {
		log.Fatal(err)
	}
}

func loadMigrationsFromDir(dir string) ([]*models.Migration, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return nil, err
	}

	var infos []models.MigrationInfo
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), ".") {
			continue
		}
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		infos = append(infos, models.MigrationInfo{
			ID:        filepath.Base(file),
			SQL:       string(content),
			Timestamp: parser.ExtractTimestampFromFilename(file),
		})
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Timestamp < infos[j].Timestamp
	})

	var migrations []*models.Migration
	for _, info := range infos {
		m, err := parser.ParseMigration(info.ID, info.SQL, info.Timestamp)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}
	return migrations, nil
}
