package main

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLI_DefaultAnalyze(t *testing.T) {
	out := runCLI(t, "CREATE TABLE users (id INT);")

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	analysisResult := result["analysisResult"].(map[string]interface{})
	migrations := analysisResult["migrations"].([]interface{})
	if len(migrations) != 1 {
		t.Errorf("expected 1 migration, got %d", len(migrations))
	}
}

func TestCLI_Analyze_NoWarnings(t *testing.T) {
	dir := testdataPath(t, "migrations")
	out := runCLI(t, dir)

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	analysisResult := result["analysisResult"].(map[string]interface{})
	migrations := analysisResult["migrations"].([]interface{})
	if len(migrations) != 2 {
		t.Errorf("expected 2 migrations, got %d", len(migrations))
	}

	warnings := analysisResult["warnings"].([]interface{})
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings for non-partitioned tables, got %d", len(warnings))
	}
}

func TestCLI_Analyze_WithWarning(t *testing.T) {
	dir := testdataPath(t, "partitioned")
	out := runCLI(t, dir)

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	analysisResult := result["analysisResult"].(map[string]interface{})
	warnings := analysisResult["warnings"].([]interface{})
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning for partitioned table ALTER, got %d", len(warnings))
	}

	warning := warnings[0].(map[string]interface{})
	if warning["type"] != "ACCESS_EXCLUSIVE_LOCK" {
		t.Errorf("expected ACCESS_EXCLUSIVE_LOCK warning, got %s", warning["type"])
	}
	if warning["tableName"] != "events" {
		t.Errorf("expected tableName 'events', got %s", warning["tableName"])
	}
}

func TestCLI_InvalidSQL(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "NOT VALID SQL AT ALL")
	cmd.Dir = cliDir(t)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error for invalid SQL")
	}
	if !strings.Contains(string(output), "error") {
		t.Errorf("expected error message, got: %s", output)
	}
}

func TestCLI_RunFlag(t *testing.T) {
	out := runCLI(t, "--run", `CREATE TABLE test_run (id INT);`)

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	// --run returns {runResult: {success, message, migrationsApplied}}
	runResult := result["runResult"].(map[string]interface{})
	if runResult["success"] != true {
		t.Errorf("expected success=true, got %v: %s", runResult["success"], runResult["message"])
	}
	if runResult["migrationsApplied"].(float64) != 1 {
		t.Errorf("expected 1 migration applied, got %v", runResult["migrationsApplied"])
	}
}

func TestCLI_RunFlag_Directory(t *testing.T) {
	dir := testdataPath(t, "migrations")
	out := runCLI(t, "--run", dir)

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	// --run returns {runResult: {success, message, migrationsApplied}}
	runResult := result["runResult"].(map[string]interface{})
	if runResult["success"] != true {
		t.Errorf("expected success=true, got %v: %s", runResult["success"], runResult["message"])
	}
	if runResult["migrationsApplied"].(float64) != 2 {
		t.Errorf("expected 2 migrations applied, got %v", runResult["migrationsApplied"])
	}
}

func TestCLI_RunWithAnalysis(t *testing.T) {
	dir := testdataPath(t, "migrations")
	out := runCLI(t, "--run", dir)

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	// Should have analysisResult with migrations inside (analyze is implicit)
	analysisResult := result["analysisResult"].(map[string]interface{})
	migrations := analysisResult["migrations"].([]interface{})
	if len(migrations) != 2 {
		t.Errorf("expected 2 migrations in analysisResult, got %d", len(migrations))
	}

	// Should have runResult
	runResult := result["runResult"].(map[string]interface{})
	if runResult["success"] != true {
		t.Errorf("expected runResult.success=true, got %v", runResult["success"])
	}
	if runResult["migrationsApplied"].(float64) != 2 {
		t.Errorf("expected 2 migrations applied, got %v", runResult["migrationsApplied"])
	}
}

func runCLI(t *testing.T, args ...string) []byte {
	t.Helper()
	cmdArgs := append([]string{"run", "."}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = cliDir(t)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Fatalf("CLI failed: %v\nstderr: %s", err, exitErr.Stderr)
		}
		t.Fatalf("CLI failed: %v", err)
	}
	return output
}

func cliDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get current file path")
	}
	return filepath.Dir(file)
}

func testdataPath(t *testing.T, subdir string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get current file path")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "testdata", subdir)
}
