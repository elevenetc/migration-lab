// Command ci orchestrates runtime analysis from the reusable GitHub workflow.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type configuration struct {
	repository string
	source     string
	directory  string
	output     string
}

func main() {
	config := configuration{
		repository: os.Getenv("CALLER_ROOT"),
		source:     os.Getenv("MIGRATION_LAB_ROOT"),
		directory:  os.Getenv("MIGRATIONS_DIRECTORY"),
		output:     os.Getenv("RESULTS_DIR"),
	}
	code, err := runCommand(context.Background(), os.Args[1:], config)
	if err != nil {
		message := fmt.Sprintf("Migration Lab automation failed: %v\n", err)
		logger := log.New(os.Stderr, "", 0)
		logger.Print(message)
		if config.output != "" {
			if writeErr := os.WriteFile(filepath.Join(config.output, "automation-error.txt"), []byte(message), 0o644); writeErr != nil {
				logger.Printf("Could not save automation error: %v", writeErr)
			}
		}
		code = 1
	}
	os.Exit(code)
}

func runCommand(ctx context.Context, args []string, config configuration) (int, error) {
	if config.output == "" {
		return 1, fmt.Errorf("RESULTS_DIR is required")
	}
	if err := os.MkdirAll(config.output, 0o755); err != nil {
		return 1, err
	}
	if config.repository == "" || config.source == "" {
		return 1, fmt.Errorf("CALLER_ROOT and MIGRATION_LAB_ROOT are required")
	}
	if len(args) != 1 {
		return 1, fmt.Errorf("usage: migration-lab-ci <plan|run>")
	}
	switch args[0] {
	case "plan":
		var event workflowEvent
		if err := readJSON(os.Getenv("GITHUB_EVENT_PATH"), &event); err != nil {
			return 1, err
		}
		plan, err := planMigrations(config.repository, config.directory, os.Getenv("GITHUB_EVENT_NAME"), event)
		if err != nil {
			return 1, err
		}
		sha, err := gitOutput(config.source, "rev-parse", "HEAD")
		if err != nil {
			return 1, err
		}
		plan.AnalyzerSHA = strings.TrimSpace(string(sha))
		if err := writeJSON(filepath.Join(config.output, "plan.json"), plan); err != nil {
			return 1, err
		}
		output, err := os.OpenFile(os.Getenv("GITHUB_OUTPUT"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return 1, err
		}
		_, writeErr := fmt.Fprintf(output, "has-targets=%t\n", len(plan.Targets) > 0)
		if err := errors.Join(writeErr, output.Close()); err != nil {
			return 1, err
		}
		if err := json.NewEncoder(os.Stdout).Encode(plan); err != nil {
			return 1, err
		}
		if len(plan.ExistingChanges) > 0 {
			if _, err := fmt.Println("Existing migration edits, deletions, and renames are recorded but not measured."); err != nil {
				return 1, err
			}
		}
		if len(plan.Targets) == 0 {
			if _, err := fmt.Println("No new migrations; skipping runtime analysis."); err != nil {
				return 1, err
			}
		}
		return 0, nil
	case "run":
		var plan migrationPlan
		if err := readJSON(filepath.Join(config.output, "plan.json"), &plan); err != nil {
			return 1, err
		}
		return runAnalysis(ctx, filepath.Join(config.source, "build", "migration-lab"), config.repository, plan, config.output)
	default:
		return 1, fmt.Errorf("unknown command %q; expected plan or run", args[0])
	}
}
