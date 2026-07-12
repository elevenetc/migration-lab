package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"migration-timeline/backend/internal/database"
	"migration-timeline/backend/internal/datasets"
	"migration-timeline/backend/internal/models"
)

// rawResponse is used for testing JSON structure without full deserialization
type rawResponse struct {
	Timeline       []json.RawMessage          `json:"timeline"`
	Map            map[string]json.RawMessage `json:"map"`
	CreateTableMap map[string]json.RawMessage `json:"createTableMap"`
	Analysis       json.RawMessage            `json:"analysis"`
}

type fakeStore struct {
	migrations []models.Migration
	err        error
}

func (f fakeStore) Migrations(string) ([]models.Migration, error) {
	return f.migrations, f.err
}

type fakeRunner struct {
	result models.RunMigrationsResult
	err    error
}

func (f fakeRunner) Run(context.Context, string) (models.RunMigrationsResult, error) {
	return f.result, f.err
}

func datasetsDB() *database.Database {
	return database.New(datasets.Datasets())
}

func TestMigrationsEndpoint(t *testing.T) {
	e := New(Config{Port: 8081, Store: datasetsDB()})

	req := httptest.NewRequest(http.MethodGet, "/api/migrations?migrationId=ecommerce", nil)
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
	store := fakeStore{migrations: []models.Migration{
		{
			ID:        "v1",
			Version:   "1",
			Timestamp: 1,
			Statements: []models.Statement{
				{
					Index: 0,
					Kind:  "CREATE_TABLE",
					SQL:   "CREATE TABLE test (id serial PRIMARY KEY)",
					Operations: []models.Operation{
						models.CreateTable{
							TableName: "test",
							Columns: []models.Column{
								{Name: "id", Type: "serial", Constraints: []string{"PRIMARY KEY"}},
							},
						},
					},
				},
			},
		},
	}}

	e := New(Config{Port: 8081, Store: store})

	req := httptest.NewRequest(http.MethodGet, "/api/migrations?migrationId=test", nil)
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

func TestMissingMigrationIdReturns404(t *testing.T) {
	e := New(Config{Port: 8081, Store: datasetsDB()})

	req := httptest.NewRequest(http.MethodGet, "/api/migrations", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d for missing migrationId, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestUnknownMigrationIdReturns404(t *testing.T) {
	e := New(Config{Port: 8081, Store: datasetsDB()})

	req := httptest.NewRequest(http.MethodGet, "/api/migrations?migrationId=bogus", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d for unknown migrationId, got %d", http.StatusNotFound, rec.Code)
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

func TestRunMigrationsUnknownIdReturns404(t *testing.T) {
	e := New(Config{Port: 8081, Runner: fakeRunner{err: models.ErrNotFound}})

	req := httptest.NewRequest(http.MethodPost, "/api/migrations/run?migrationId=bogus", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d for unknown migrationId, got %d", http.StatusNotFound, rec.Code)
	}
}
