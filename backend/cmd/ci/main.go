// Command ci orchestrates runtime analysis from the reusable GitHub workflow.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"migration-lab/backend/internal/ci"
)

func main() {
	config := ci.Config{
		Repository:       os.Getenv("CALLER_ROOT"),
		Source:           os.Getenv("MIGRATION_LAB_ROOT"),
		Directory:        os.Getenv("MIGRATIONS_DIRECTORY"),
		Output:           os.Getenv("RESULTS_DIR"),
		EventName:        os.Getenv("GITHUB_EVENT_NAME"),
		EventPath:        os.Getenv("GITHUB_EVENT_PATH"),
		GitHubOutput:     os.Getenv("GITHUB_OUTPUT"),
		GitHubRepository: os.Getenv("GITHUB_REPOSITORY"),
		GitHubRunID:      os.Getenv("GITHUB_RUN_ID"),
		GitHubRunAttempt: os.Getenv("GITHUB_RUN_ATTEMPT"),
		AnalysisStatus:   os.Getenv("ANALYSIS_STATUS"),
		GitHubToken:      os.Getenv("GH_TOKEN"),
	}
	code, err := ci.RunCommand(context.Background(), os.Args[1:], config)
	if err != nil {
		message := fmt.Sprintf("Migration Lab automation failed: %v\n", err)
		logger := log.New(os.Stderr, "", 0)
		logger.Print(message)
		if config.Output != "" {
			if writeErr := os.WriteFile(filepath.Join(config.Output, "automation-error.txt"), []byte(message), 0o644); writeErr != nil {
				logger.Printf("Could not save automation error: %v", writeErr)
			}
		}
		code = 1
	}
	os.Exit(code)
}
