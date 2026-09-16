package ci

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config supplies the workspace paths and GitHub workflow environment.
type Config struct {
	Repository       string
	Source           string
	Directory        string
	Output           string
	EventName        string
	EventPath        string
	GitHubOutput     string
	GitHubRepository string
	GitHubRunID      string
	GitHubRunAttempt string
	AnalysisStatus   string
	GitHubToken      string
}

// RunCommand executes the plan, run, or comment workflow step.
func RunCommand(ctx context.Context, args []string, config Config) (int, error) {
	if config.Output == "" {
		return 1, fmt.Errorf("RESULTS_DIR is required")
	}
	if err := os.MkdirAll(config.Output, 0o755); err != nil {
		return 1, err
	}
	if len(args) != 1 {
		return 1, fmt.Errorf("usage: migration-lab-ci <plan|run|comment>")
	}
	if args[0] == "comment" {
		return 0, commentOnPR(ctx, config)
	}
	if config.Repository == "" || config.Source == "" {
		return 1, fmt.Errorf("CALLER_ROOT and MIGRATION_LAB_ROOT are required")
	}
	switch args[0] {
	case "plan":
		var event workflowEvent
		if err := readJSON(config.EventPath, &event); err != nil {
			return 1, err
		}
		plan, err := planMigrations(config.Repository, config.Directory, config.EventName, event)
		if err != nil {
			return 1, err
		}
		sha, err := gitOutput(config.Source, "rev-parse", "HEAD")
		if err != nil {
			return 1, err
		}
		plan.AnalyzerSHA = strings.TrimSpace(string(sha))
		if err := writeJSON(filepath.Join(config.Output, "plan.json"), plan); err != nil {
			return 1, err
		}
		output, err := os.OpenFile(config.GitHubOutput, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
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
		if err := readJSON(filepath.Join(config.Output, "plan.json"), &plan); err != nil {
			return 1, err
		}
		return runAnalysis(ctx, filepath.Join(config.Source, "build", "migration-lab"), config.Repository, plan, config.Output)
	default:
		return 1, fmt.Errorf("unknown command %q; expected plan, run, or comment", args[0])
	}
}
