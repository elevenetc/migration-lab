package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"migration-lab/backend/internal/models"
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

func TestPrintJSONReportsWriteFailure(t *testing.T) {
	// A read-only file makes stdout writes fail without relying on platform-
	// specific devices or broken-pipe signal handling.
	output, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	os.Stdout = output
	t.Cleanup(func() {
		os.Stdout = previous
		if err := output.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := printJSON(map[string]bool{"ok": true}); err == nil {
		t.Fatal("expected JSON output failure to be returned")
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

func TestCLI_RuntimeFlag(t *testing.T) {
	dir := testdataPath(t, "migrations")
	out := runCLI(t, "--runtime", "--rows", "1000", dir)

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	runtimeResult := result["runtimeResult"].(map[string]interface{})
	if runtimeResult["verdict"] != "COMPLETED" {
		t.Errorf("expected verdict COMPLETED, got %v: %s", runtimeResult["verdict"], runtimeResult["message"])
	}

	seeded := runtimeResult["seeded"].([]interface{})
	if len(seeded) != 1 {
		t.Fatalf("expected the touched table to be seeded, got %v", seeded)
	}
	users := seeded[0].(map[string]interface{})
	if users["table"] != "users" || users["rows"].(float64) != 1000 {
		t.Errorf("expected users seeded to 1000 rows, got %v", users)
	}
}

func TestCLI_RuntimeFlag_SelectsMigrationAndIgnoresLaterSQL(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"V1__create.sql": "CREATE TABLE users (id INT);",
		"V2__add.sql":    "ALTER TABLE users ADD COLUMN email TEXT;",
		"V3__later.sql":  "NOT VALID SQL;",
	}
	for name, sql := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(sql), 0600); err != nil {
			t.Fatal(err)
		}
	}
	out := runCLI(t, "--runtime", "--migration", "V2__add.sql", "--rows", "100", dir)
	var result struct {
		RuntimeResult  models.RuntimeAnalysisResult `json:"runtimeResult"`
		AnalysisResult struct {
			Migrations []json.RawMessage `json:"migrations"`
		} `json:"analysisResult"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatal(err)
	}
	if result.RuntimeResult.MigrationID != "V2__add.sql" || result.RuntimeResult.Verdict != "COMPLETED" {
		t.Fatalf("unexpected runtime result: %+v", result.RuntimeResult)
	}
	if len(result.RuntimeResult.Seeded) != 1 || result.RuntimeResult.Seeded[0].Rows != 100 {
		t.Fatalf("expected the predecessor's table to be seeded: %+v", result.RuntimeResult.Seeded)
	}
	if len(result.AnalysisResult.Migrations) != 2 {
		t.Fatalf("expected static analysis through the target, got %d migrations", len(result.AnalysisResult.Migrations))
	}
}

func TestCLI_MigrationFlagValidation(t *testing.T) {
	for _, test := range []struct {
		name string
		args []string
		want string
	}{
		{"requires runtime", []string{"--migration", "V2__missing.sql"}, "--migration requires --runtime"},
		{"unknown target", []string{"--runtime", "--migration", "V2__missing.sql"}, `migration "V2__missing.sql" not found`},
		{"report conflict", []string{"--runtime", "--report=report.html"}, "none of the others can be"},
	} {
		t.Run(test.name, func(t *testing.T) {
			stderr := runCLIExpectingFailure(t, append(test.args, testdataPath(t, "migrations"))...)
			if !strings.Contains(stderr, test.want) {
				t.Fatalf("expected %q, got %s", test.want, stderr)
			}
		})
	}
}

func TestCLI_RuntimeFlag_ReportsATableItCannotSeed(t *testing.T) {
	// The partitioned parent has no partitions yet, so no row can be routed into it.
	dir := testdataPath(t, "partitioned")
	out := runCLI(t, "--runtime", "--rows", "1000", dir)

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	runtimeResult := result["runtimeResult"].(map[string]interface{})
	if !hasFindingOfType(runtimeResult, "SEED_FAILED") {
		t.Errorf("expected a SEED_FAILED finding, got %v", runtimeResult["findings"])
	}
	if runtimeResult["verdict"] != "COMPLETED" {
		t.Errorf("expected the run to continue past a failed seed, got %v", runtimeResult["verdict"])
	}
}

func TestCLI_RuntimeFlag_FailsTheBuildWhenTheDeadlineIsMissed(t *testing.T) {
	dir := testdataPath(t, "rewrite")
	stderr := runCLIExpectingFailure(t, "--runtime", "--rows", "200000", "--deadline-ms", "1", dir)

	if !strings.Contains(stderr, "EXCEEDS_DEADLINE") {
		t.Errorf("expected the exit message to name the verdict, got: %s", stderr)
	}
}

func hasFindingOfType(runtimeResult map[string]interface{}, findingType string) bool {
	for _, finding := range runtimeResult["findings"].([]interface{}) {
		if finding.(map[string]interface{})["type"] == findingType {
			return true
		}
	}
	return false
}

func runCLI(t *testing.T, args ...string) []byte {
	t.Helper()
	cmdArgs := append([]string{"run", "."}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = cliDir(t)
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			t.Fatalf("CLI failed: %v\nstderr: %s", err, exitErr.Stderr)
		}
		t.Fatalf("CLI failed: %v", err)
	}
	return output
}

// runCLIExpectingFailure returns stderr of a run that must exit non-zero, which
// is how the runtime pass fails the CI job it runs in.
func runCLIExpectingFailure(t *testing.T, args ...string) string {
	t.Helper()
	cmdArgs := append([]string{"run", "."}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = cliDir(t)

	output, err := cmd.Output()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected a non-zero exit, got err %v and output: %s", err, output)
	}
	return string(exitErr.Stderr)
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
