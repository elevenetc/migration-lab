package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"testing"
	"time"

	"migration-lab/backend/internal/analysis/runtime"
	"migration-lab/backend/internal/database"
	"migration-lab/backend/internal/datasets"
	"migration-lab/backend/internal/models"
)

// rawResponse is used for testing JSON structure without full deserialization
type rawResponse struct {
	Timeline       []json.RawMessage          `json:"timeline"`
	Map            map[string]json.RawMessage `json:"map"`
	CreateTableMap map[string]json.RawMessage `json:"createTableMap"`
	Analysis       json.RawMessage            `json:"analysis"`
}

type fakeStore struct {
	infos []models.MigrationInfo
	err   error
}

func (f fakeStore) MigrationInfos(string) ([]models.MigrationInfo, error) {
	return f.infos, f.err
}

func (f fakeStore) DatasetIds() []string {
	return nil
}

func fakeRun(result models.RunMigrationsResult, err error) RunDatasetMigrations {
	return func(context.Context, string) (models.RunMigrationsResult, error) {
		return result, err
	}
}

// captureAnalyse stands in for runtime.Analyse, recording the request the handler
// built so the query-parameter mapping is testable without Docker.
func captureAnalyse(result models.RuntimeAnalysisResult, err error, seen *runtime.Request) RuntimeAnalyse {
	return func(_ context.Context, request runtime.Request) (models.RuntimeAnalysisResult, error) {
		*seen = request
		return result, err
	}
}

func datasetsDB() *database.Database {
	return database.New(datasets.Datasets())
}

func TestMigrationsEndpoint(t *testing.T) {
	e := New(Config{Port: 8081, Store: datasetsDB()})

	req := httptest.NewRequest(http.MethodGet, "/api/migrations?datasetId=ecommerce", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response rawResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	want := len(datasets.Datasets()["ecommerce"])

	if len(response.Timeline) != want {
		t.Errorf("expected %d migrations in timeline, got %d", want, len(response.Timeline))
	}

	if len(response.Map) != want {
		t.Errorf("expected %d entries in map, got %d", want, len(response.Map))
	}

	if _, ok := response.Map["v1"]; !ok {
		t.Error("expected v1 in migration map")
	}

	if _, ok := response.CreateTableMap["users"]; !ok {
		t.Error("expected users in createTableMap")
	}

	if response.Analysis == nil {
		t.Error("expected analysis to be present")
	}
}

func TestResponseStructure(t *testing.T) {
	store := fakeStore{infos: []models.MigrationInfo{
		{ID: "v1", SQL: "CREATE TABLE test (id serial PRIMARY KEY)", Timestamp: 1},
	}}

	e := New(Config{Port: 8081, Store: store})

	req := httptest.NewRequest(http.MethodGet, "/api/migrations?datasetId=test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	requiredFields := []string{"timeline", "map", "createTableMap", "analysis"}
	for _, field := range requiredFields {
		if _, ok := raw[field]; !ok {
			t.Errorf("expected field %q in response", field)
		}
	}
}

func TestDatasetsEndpoint(t *testing.T) {
	e := New(Config{Port: 8081, Store: datasetsDB()})

	req := httptest.NewRequest(http.MethodGet, "/api/datasets", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response struct {
		Datasets []string `json:"datasets"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(response.Datasets) != len(datasets.Datasets()) {
		t.Errorf("expected %d datasets, got %d", len(datasets.Datasets()), len(response.Datasets))
	}

	if !sort.StringsAreSorted(response.Datasets) {
		t.Errorf("expected datasets sorted, got %v", response.Datasets)
	}

	if !slices.Contains(response.Datasets, "ecommerce") {
		t.Errorf("expected datasets to contain ecommerce, got %v", response.Datasets)
	}
}

func TestMissingMigrationIdReturns404(t *testing.T) {
	e := New(Config{Port: 8081, Store: datasetsDB()})

	req := httptest.NewRequest(http.MethodGet, "/api/migrations", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d for missing datasetId, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestUnknownMigrationIdReturns404(t *testing.T) {
	e := New(Config{Port: 8081, Store: datasetsDB()})

	req := httptest.NewRequest(http.MethodGet, "/api/migrations?datasetId=bogus", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d for unknown datasetId, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestMigrationsPathLoadsFromDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "V1__create_users.sql", "CREATE TABLE users (id INT);")
	writeFile(t, dir, "V2__add_email.sql", "ALTER TABLE users ADD COLUMN email TEXT;")
	writeFile(t, dir, ".ignore.sql", "CREATE TABLE ignored (id INT);")

	e := New(Config{Port: 8081, Store: fakeStore{err: models.ErrNotFound}})

	target := "/api/migrations?migrationsPath=" + url.QueryEscape(dir)
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d (%s)", http.StatusOK, rec.Code, rec.Body.String())
	}

	var response rawResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(response.Timeline) != 2 {
		t.Errorf("expected 2 migrations (dotfile skipped), got %d", len(response.Timeline))
	}
	if _, ok := response.CreateTableMap["users"]; !ok {
		t.Error("expected users in createTableMap")
	}
}

func TestMigrationsPathNonExistentReturns400(t *testing.T) {
	e := New(Config{Port: 8081, Store: datasetsDB()})

	req := httptest.NewRequest(http.MethodGet, "/api/migrations?migrationsPath=/no/such/dir", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for bad path, got %d", http.StatusBadRequest, rec.Code)
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", name, err)
	}
}

func TestCORSHeaders(t *testing.T) {
	e := New(Config{Port: 8081, Store: fakeStore{}})

	req := httptest.NewRequest(http.MethodOptions, "/api/migrations", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Error("expected CORS header for localhost:3000")
	}
}

// An unknown dataset is refused while resolving the timeline, so no container is
// ever started for it.
func TestRuntimeUnknownIdReturns404(t *testing.T) {
	var seen runtime.Request
	e := New(Config{
		Port:    8081,
		Store:   fakeStore{err: models.ErrNotFound},
		Analyse: captureAnalyse(models.RuntimeAnalysisResult{}, nil, &seen),
	})

	req := httptest.NewRequest(http.MethodPost, "/api/migrations/runtime-analysis?datasetId=bogus", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d for unknown datasetId, got %d", http.StatusNotFound, rec.Code)
	}
	if seen.Migrations != nil {
		t.Error("expected the analysis never to be reached for an unknown dataset")
	}
}

func TestRuntimePassesTheQueryThroughAndReturnsTheResult(t *testing.T) {
	var seen runtime.Request
	store := fakeStore{infos: []models.MigrationInfo{
		{ID: "v1", SQL: "CREATE TABLE t (id int);", Timestamp: 1},
		{ID: "v3", SQL: "ALTER TABLE t ADD COLUMN a int;", Timestamp: 3},
	}}
	e := New(Config{
		Port:  8081,
		Store: store,
		Analyse: captureAnalyse(models.RuntimeAnalysisResult{Migrations: []models.RuntimeMigrationResult{{
			MigrationID: "v3",
			Verdict:     models.RuntimeExceedsDeadline,
			Retry:       models.RetryFailureLoop,
		}}}, nil, &seen),
	})

	target := "/api/migrations/runtime-analysis?datasetId=events&migration=v3&rows=5000&deadlineMs=1500"
	req := httptest.NewRequest(http.MethodPost, target, nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d (%s)", http.StatusOK, rec.Code, rec.Body.String())
	}

	if seen.TargetMigrationID != "v3" || seen.Rows != 5000 || seen.Deadline != 1500*time.Millisecond {
		t.Errorf("analysis received target %q, rows %d, deadline %s", seen.TargetMigrationID, seen.Rows, seen.Deadline)
	}
	if len(seen.Migrations) != 2 || seen.Migrations[1].SQL != "ALTER TABLE t ADD COLUMN a int;" {
		t.Errorf("expected the dataset's raw migrations to be handed over, got %+v", seen.Migrations)
	}

	var analysis models.RuntimeAnalysisResult
	if err := json.Unmarshal(rec.Body.Bytes(), &analysis); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if len(analysis.Migrations) != 1 {
		t.Fatalf("expected the target's measurement, got %+v", analysis)
	}
	if result := analysis.Migrations[0]; result.Verdict != models.RuntimeExceedsDeadline || result.Retry != models.RetryFailureLoop {
		t.Errorf("expected the measured verdicts to be returned, got %+v", result)
	}
}

// A mistyped measurement parameter must be refused, not silently replaced by the
// default — the caller would otherwise get a different measurement than the one
// it asked for, with nothing saying so.
func TestRuntimeRejectsANonNumericParameter(t *testing.T) {
	for _, query := range []string{"rows=abc", "deadlineMs=5oo"} {
		var seen runtime.Request
		e := New(Config{
			Port:    8081,
			Store:   fakeStore{infos: []models.MigrationInfo{{ID: "v1", SQL: "CREATE TABLE t (id int);"}}},
			Analyse: captureAnalyse(models.RuntimeAnalysisResult{}, nil, &seen),
		})

		req := httptest.NewRequest(http.MethodPost, "/api/migrations/runtime-analysis?datasetId=x&"+query, nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: expected status %d, got %d", query, http.StatusBadRequest, rec.Code)
		}
		if seen.Migrations != nil {
			t.Errorf("%s: expected no container to be started", query)
		}
	}
}

func TestRuntimeFromAPathResolvesTheDirectory(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "V1__create_users.sql", "CREATE TABLE users (id INT);")
	writeFile(t, dir, "V2__add_email.sql", "ALTER TABLE users ADD COLUMN email TEXT;")

	var seen runtime.Request
	e := New(Config{
		Port:    8081,
		Store:   fakeStore{err: models.ErrNotFound},
		Analyse: captureAnalyse(models.RuntimeAnalysisResult{}, nil, &seen),
	})

	target := "/api/migrations/runtime-analysis?migrationsPath=" + url.QueryEscape(dir)
	req := httptest.NewRequest(http.MethodPost, target, nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d (%s)", http.StatusOK, rec.Code, rec.Body.String())
	}
	if len(seen.Migrations) != 2 {
		t.Errorf("expected the directory's migrations to be handed over, got %+v", seen.Migrations)
	}
	// No `migration` parameter: the analysis defaults to the newest.
	if seen.TargetMigrationID != "" {
		t.Errorf("expected no explicit target, got %q", seen.TargetMigrationID)
	}
}

func TestRunMigrationsUnknownIdReturns404(t *testing.T) {
	e := New(Config{Port: 8081, Run: fakeRun(models.RunMigrationsResult{}, models.ErrNotFound)})

	req := httptest.NewRequest(http.MethodPost, "/api/migrations/run?datasetId=bogus", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d for unknown datasetId, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestRuntimeWithoutAMigrationMeasuresEveryMigration(t *testing.T) {
	var seen runtime.Request
	store := fakeStore{infos: []models.MigrationInfo{
		{ID: "v1", SQL: "CREATE TABLE t (id int);", Timestamp: 1},
		{ID: "v2", SQL: "ALTER TABLE t ADD COLUMN a int;", Timestamp: 2},
	}}
	e := New(Config{
		Port:  8081,
		Store: store,
		Analyse: captureAnalyse(models.RuntimeAnalysisResult{
			Migrations: []models.RuntimeMigrationResult{{MigrationID: "v1"}, {MigrationID: "v2"}},
			Message:    "all 2 migrations completed in 3 ms",
		}, nil, &seen),
	})

	req := httptest.NewRequest(http.MethodPost, "/api/migrations/runtime-analysis?datasetId=events", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d (%s)", http.StatusOK, rec.Code, rec.Body.String())
	}
	if seen.TargetMigrationID != "" || len(seen.Migrations) != 2 {
		t.Errorf("expected no target and every migration handed over, got target %q and %d migrations", seen.TargetMigrationID, len(seen.Migrations))
	}

	var result models.RuntimeAnalysisResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if len(result.Migrations) != 2 {
		t.Errorf("expected both measurements returned, got %+v", result)
	}
}
