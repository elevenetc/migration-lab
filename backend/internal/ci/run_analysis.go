package ci

import (
	"context"
	"fmt"
	"path/filepath"
)

type migrationResult struct {
	Migration string `json:"migration"`
	ExitCode  int    `json:"exitCode"`
	Result    string `json:"result"`
	Log       string `json:"log"`
}

func runAnalysis(ctx context.Context, binary, repository string, plan migrationPlan, output string) (int, error) {
	if len(plan.Targets) == 0 {
		return 0, nil
	}
	directory := filepath.Join(repository, filepath.FromSlash(plan.Directory))
	code, err := runCLI(ctx, binary, []string{directory}, output, "static-analysis")
	if err != nil {
		return 1, err
	}
	if code != 0 {
		return 1, fmt.Errorf("could not parse migration history; see static-analysis.stderr.log")
	}
	results := make([]migrationResult, 0, len(plan.Targets))
	verdict := 0
	for i, target := range plan.Targets {
		if _, err := fmt.Printf("Analyzing %s (%d/%d)\n", target, i+1, len(plan.Targets)); err != nil {
			return 1, err
		}
		name := filepath.Base(target)
		code, err := runCLI(ctx, binary, []string{"--runtime", "--migration", target, directory}, output, name)
		if err != nil {
			return 1, err
		}
		results = append(results, migrationResult{Migration: target, ExitCode: code, Result: name + ".json", Log: name + ".stderr.log"})
		// Persist each result so a later failure does not discard earlier evidence.
		if err := writeJSON(filepath.Join(output, "results.json"), results); err != nil {
			return 1, err
		}
		if code != 0 {
			verdict = 1
		}
	}
	return verdict, nil
}
