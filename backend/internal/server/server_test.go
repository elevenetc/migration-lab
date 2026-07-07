package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"migration-timeline/backend/internal/models"
	"migration-timeline/backend/internal/server/dummy"
)

// rawResponse is used for testing JSON structure without full deserialization
type rawResponse struct {
	Timeline       []json.RawMessage          `json:"timeline"`
	Map            map[string]json.RawMessage `json:"map"`
	CreateTableMap map[string]json.RawMessage `json:"createTableMap"`
	Analysis       json.RawMessage            `json:"analysis"`
}

func TestMigrationsEndpoint(t *testing.T) {
	cfg := Config{
		Port:               8081,
		MigrationsProvider: dummy.GetComplexEcommerceMigrations,
	}
	e := New(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/migrations", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response rawResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(response.Timeline) != 29 {
		t.Errorf("expected 29 migrations in timeline, got %d", len(response.Timeline))
	}

	if len(response.Map) != 29 {
		t.Errorf("expected 29 entries in map, got %d", len(response.Map))
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
	provider := func() ([]*models.Migration, error) {
		return []*models.Migration{
			{
				ID:        "v1",
				Version:   "1",
				Timestamp: 1,
				Operations: []models.Operation{
					models.CreateTable{
						MigrationID: "v1",
						TableName:   "test",
						Columns: []models.Column{
							{Name: "id", Type: "serial", Constraints: []string{"PRIMARY KEY"}},
						},
					},
				},
			},
		}, nil
	}

	cfg := Config{
		Port:               8081,
		MigrationsProvider: provider,
	}
	e := New(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/migrations", nil)
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

func TestCORSHeaders(t *testing.T) {
	cfg := Config{
		Port:               8081,
		MigrationsProvider: func() ([]*models.Migration, error) { return nil, nil },
	}
	e := New(cfg)

	req := httptest.NewRequest(http.MethodOptions, "/api/migrations", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Error("expected CORS header for localhost:3000")
	}
}

func TestRunMigrationsEndpointRegistered(t *testing.T) {
	cfg := Config{
		Port:                  8081,
		MigrationsProvider:    func() ([]*models.Migration, error) { return nil, nil },
		MigrationsRunProvider: dummy.GetComplexEcommerceMigrationsForRunner,
	}
	e := New(cfg)

	req := httptest.NewRequest(http.MethodPost, "/api/migrations/run", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Should return OK, not 404 (route not found)
	if rec.Code == http.StatusNotFound {
		t.Error("expected /api/migrations/run endpoint to be registered")
	}
}
