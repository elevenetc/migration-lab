package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"migration-timeline/backend/internal/database"
	"migration-timeline/backend/internal/dummy"
	"migration-timeline/backend/internal/models"
	"migration-timeline/backend/internal/parser"
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

	datasets := dummy.Datasets()

	if *migrationsDir != "" {
		infos, err := loadMigrationInfosFromDir(*migrationsDir)
		if err != nil {
			log.Fatal(err)
		}
		datasets[filepath.Base(*migrationsDir)] = infos
	}

	db := database.New(datasets)
	cfg := server.Config{
		Port:   *port,
		Store:  db,
		Runner: runner.MigrationRunner{Store: db},
	}
	return cfg
}

func loadMigrationInfosFromDir(dir string) ([]models.MigrationInfo, error) {
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

	return infos, nil
}
