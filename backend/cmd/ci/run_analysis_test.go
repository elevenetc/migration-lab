package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// Re-execute the test binary as a CLI stub to test real process exit codes,
// argument boundaries, cancellation, and file output without another language.
func TestMain(m *testing.M) {
	if os.Getenv("MIGRATION_LAB_TEST_CLI") == "1" {
		args := os.Args[1:]
		if _, err := fmt.Fprintln(os.Stderr, "CLI diagnostic"); err != nil {
			os.Exit(3)
		}
		if slices.Contains(args, "wait") {
			if _, err := fmt.Print("partial output"); err != nil {
				os.Exit(3)
			}
			time.Sleep(time.Minute)
		}
		if os.Getenv("MIGRATION_LAB_TEST_STATIC_FAILURE") == "1" {
			os.Exit(2)
		}
		if index := slices.Index(args, "--migration"); index >= 0 {
			target := args[index+1]
			if err := json.NewEncoder(os.Stdout).Encode(map[string]any{"runtimeResult": map[string]string{"migrationId": target}}); err != nil {
				os.Exit(3)
			}
			if strings.HasPrefix(target, "V2__") {
				os.Exit(1)
			}
		} else {
			if _, err := fmt.Println(`{"analysisResult":{"migrations":[]}}`); err != nil {
				os.Exit(3)
			}
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestRunAnalysisPreservesFailureAndContinues(t *testing.T) {
	binary := stubCLI(t)
	output := t.TempDir()
	plan := migrationPlan{Directory: "db", Targets: []string{"V2__space $(name).sql", "V10__last.sql"}}
	code, err := runAnalysis(context.Background(), binary, t.TempDir(), plan, output)
	if err != nil || code != 1 {
		t.Fatalf("unexpected outcome: code %d, %v", code, err)
	}
	var results []migrationResult
	if err := readJSON(filepath.Join(output, "results.json"), &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 || results[0].ExitCode != 1 || results[1].ExitCode != 0 {
		t.Fatalf("expected failure followed by success: %+v", results)
	}
	for i, result := range results {
		var data struct {
			RuntimeResult struct {
				MigrationID string `json:"migrationId"`
			} `json:"runtimeResult"`
		}
		if err := readJSON(filepath.Join(output, result.Result), &data); err != nil {
			t.Fatal(err)
		}
		if result.Migration != plan.Targets[i] || data.RuntimeResult.MigrationID != plan.Targets[i] {
			t.Fatalf("target was altered or reordered: %+v", result)
		}
		if log, err := os.ReadFile(filepath.Join(output, result.Log)); err != nil || !strings.Contains(string(log), "diagnostic") {
			t.Fatalf("missing diagnostics: %q, %v", log, err)
		}
	}
}

func TestRunAnalysisSkipsEmptyPlan(t *testing.T) {
	code, err := runAnalysis(context.Background(), "missing-cli", t.TempDir(), migrationPlan{}, t.TempDir())
	if code != 0 || err != nil {
		t.Fatalf("empty plan should not invoke CLI: code %d, %v", code, err)
	}
}

func TestRunAnalysisStopsAtStaticFailure(t *testing.T) {
	binary := stubCLI(t)
	t.Setenv("MIGRATION_LAB_TEST_STATIC_FAILURE", "1")
	output := t.TempDir()
	_, err := runAnalysis(context.Background(), binary, t.TempDir(), migrationPlan{Targets: []string{"V10__last.sql"}}, output)
	if err == nil || !strings.Contains(err.Error(), "parse migration history") {
		t.Fatalf("expected parsing failure, got %v", err)
	}
	if log, err := os.ReadFile(filepath.Join(output, "static-analysis.stderr.log")); err != nil || !strings.Contains(string(log), "diagnostic") {
		t.Fatalf("missing diagnostics: %q, %v", log, err)
	}
	if _, err := os.Stat(filepath.Join(output, "001.json")); !os.IsNotExist(err) {
		t.Fatal("runtime analysis should not have started")
	}
}

func TestRunCLITimeoutPreservesPartialOutput(t *testing.T) {
	binary := stubCLI(t)
	output := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	code, err := runCLI(ctx, binary, []string{"wait"}, output, "timeout")
	if err != nil || code != 124 {
		t.Fatalf("expected timeout: code %d, %v", code, err)
	}
	if out, err := os.ReadFile(filepath.Join(output, "timeout.json")); err != nil || string(out) != "partial output" {
		t.Fatalf("missing partial output: %q, %v", out, err)
	}
	if log, err := os.ReadFile(filepath.Join(output, "timeout.stderr.log")); err != nil || !strings.Contains(string(log), "timed out") {
		t.Fatalf("missing timeout diagnostic: %q, %v", log, err)
	}
}

func TestRunCommandReadsSavedPlan(t *testing.T) {
	binary := stubCLI(t)
	source := t.TempDir()
	if err := os.Mkdir(filepath.Join(source, "build"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(binary, filepath.Join(source, "build", "migration-lab")); err != nil {
		t.Fatal(err)
	}
	output := t.TempDir()
	if err := writeJSON(filepath.Join(output, "plan.json"), migrationPlan{Directory: "db", Targets: []string{"V10__last.sql"}}); err != nil {
		t.Fatal(err)
	}
	code, err := runCommand(context.Background(), []string{"run"}, configuration{repository: t.TempDir(), source: source, output: output})
	if err != nil || code != 0 {
		t.Fatalf("run command failed: code %d, %v", code, err)
	}
	var results []migrationResult
	if err := readJSON(filepath.Join(output, "results.json"), &results); err != nil || len(results) != 1 || results[0].Migration != "V10__last.sql" {
		t.Fatalf("unexpected saved results: %+v, %v", results, err)
	}
}

func stubCLI(t *testing.T) string {
	t.Helper()
	t.Setenv("MIGRATION_LAB_TEST_CLI", "1")
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return binary
}
