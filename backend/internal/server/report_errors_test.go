package server

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"migration-lab/backend/internal/monitoring"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
)

type eventTransport struct {
	events []*sentry.Event
	logs   []sentry.Log
}

func (t *eventTransport) SendEvent(event *sentry.Event) {
	if len(event.Logs) > 0 {
		t.logs = append(t.logs, event.Logs...)
		return
	}
	t.events = append(t.events, event)
}
func (*eventTransport) Configure(sentry.ClientOptions)        {}
func (*eventTransport) Flush(time.Duration) bool              { return true }
func (*eventTransport) FlushWithContext(context.Context) bool { return true }
func (*eventTransport) Close()                                {}

func testHub(t *testing.T) (*sentry.Hub, *eventTransport) {
	t.Helper()
	base, err := monitoring.NewHub("https://public@example.invalid/1", "test", "test-release")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(base.Client().Close)
	transport := &eventTransport{}
	options := base.Client().Options()
	options.Transport = transport
	client, err := sentry.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(client.Close)
	return sentry.NewHub(client, base.Scope()), transport
}

func TestRequestErrorsAreCapturedOnceWithIsolatedContext(t *testing.T) {
	hub, transport := testHub(t)
	e := New(Config{Sentry: hub, Store: fakeStore{err: errors.New("store unavailable")}})
	e.GET("/panic", func(c echo.Context) error { panic("broken handler") })
	e.GET("/error", func(c echo.Context) error { return errors.New("returned error") })

	for _, path := range []string{"/api/migrations", "/panic", "/error"} {
		req := httptest.NewRequest(http.MethodGet, path+"?migrationId=demo&migration=V2&privatePath=/private/sql", strings.NewReader("secret SQL"))
		req.Header.Set("Authorization", "secret")
		req.Header.Set("Cookie", "session=secret")
		rec := httptest.NewRecorder()
		before := len(transport.events)
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("%s: status %d", path, rec.Code)
		}
		if len(transport.events) != before+1 {
			t.Fatalf("%s: expected one event, got %d", path, len(transport.events)-before)
		}
		event := transport.events[before]
		if event.Tags["request_id"] == "" || event.Tags["request_id"] != rec.Header().Get(echo.HeaderXRequestID) {
			t.Fatalf("request ID missing: %+v", event.Tags)
		}
		if event.Tags["dataset_id"] != "demo" || event.Tags["migration_id"] != "V2" || event.Tags["service"] != "backend" {
			t.Fatalf("missing context: %+v", event.Tags)
		}
		if event.Environment != "test" || event.Release != "test-release" {
			t.Fatal("missing environment/release")
		}
		if event.Request == nil || event.Request.QueryString != "" || event.Request.Data != "" || len(event.Request.Headers) != 0 || event.Request.Cookies != "" || strings.Contains(event.Request.URL, "private") {
			t.Fatalf("request was not scrubbed: %+v", event.Request)
		}
		hasStack := len(event.Exception) > 0 && event.Exception[len(event.Exception)-1].Stacktrace != nil
		for _, thread := range event.Threads {
			hasStack = hasStack || thread.Stacktrace != nil
		}
		if !hasStack {
			t.Fatal("missing stack trace")
		}
	}
	// A following request must not inherit a previous request's migration tags.
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/error", nil))
	last := transport.events[len(transport.events)-1]
	if last.Tags["dataset_id"] != "" || last.Tags["migration_id"] != "" {
		t.Fatalf("leaked context: %+v", last.Tags)
	}
}

func TestExpectedResponsesAreNotReported(t *testing.T) {
	hub, transport := testHub(t)
	e := New(Config{Sentry: hub, Store: datasetsDB(), Runner: fakeRunner{}})
	e.GET("/bad-input", func(c echo.Context) error { return echo.NewHTTPError(400, "invalid input") })
	for _, path := range []string{"/health", "/api/migrations?migrationId=missing", "/api/migrations?migrationsPath=/missing", "/bad-input"} {
		e.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}
	// Unsuccessful SQL execution is still a normal analysis result (HTTP 200).
	e.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/migrations/run", nil))
	if len(transport.events) != 0 {
		t.Fatalf("expected no events, got %d", len(transport.events))
	}
}

func TestInfrastructureFailureReportsDespiteSuccessfulHTTPResponse(t *testing.T) {
	hub, transport := testHub(t)
	e := New(Config{Sentry: hub})
	e.GET("/setup", func(c echo.Context) error {
		monitoring.CaptureError(c.Request().Context(), errors.New("Docker unavailable"))
		monitoring.CaptureError(c.Request().Context(), context.Canceled)
		return c.JSON(http.StatusOK, map[string]bool{"success": false})
	})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/setup", nil))
	if rec.Code != 200 || len(transport.events) != 1 {
		t.Fatalf("status %d, events %d", rec.Code, len(transport.events))
	}
}

func TestPanicRecoveryWithoutSentry(t *testing.T) {
	e := New(Config{})
	e.GET("/panic", func(c echo.Context) error { panic("no DSN") })
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/panic", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestLogsAreDeliveredWithRequestContextAndKeptLocally(t *testing.T) {
	hub, transport := testHub(t)
	var local bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&local)
	t.Cleanup(func() { log.SetOutput(previous) })
	e := New(Config{Sentry: hub})
	e.GET("/logged", func(c echo.Context) error {
		monitoring.Infof(c.Request().Context(), "Applying migration %s", "V2")
		monitoring.Errorf(c.Request().Context(), "Cleanup failed: %s", "test failure")
		return errors.New("handler failed")
	})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/logged?migrationId=demo&migration=V2&migrationsPath=/private/sql", nil))
	e.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health", nil))
	monitoring.Infof(context.Background(), "CLI stays local")
	if !hub.Flush(time.Second) {
		t.Fatal("logs did not flush")
	}
	if len(transport.logs) != 4 {
		t.Fatalf("expected 4 server logs, got %+v", transport.logs)
	}
	for i, entry := range transport.logs {
		if entry.Attributes["service"].AsString() != "backend" || entry.Attributes["sentry.environment"].AsString() != "test" || entry.Attributes["sentry.release"].AsString() != "test-release" {
			t.Fatalf("missing service/environment/release: %+v", entry.Attributes)
		}
		if i < 3 {
			if entry.Attributes["request_id"].AsString() != rec.Header().Get(echo.HeaderXRequestID) || entry.Attributes["dataset_id"].AsString() != "demo" || entry.Attributes["migration_id"].AsString() != "V2" {
				t.Fatalf("missing request context: %+v", entry.Attributes)
			}
		} else if entry.Attributes["dataset_id"].AsString() != "" || entry.Attributes["migration_id"].AsString() != "" {
			t.Fatalf("leaked request context: %+v", entry.Attributes)
		}
		if strings.Contains(entry.Body, "private") || !strings.Contains(local.String(), entry.Body) {
			t.Fatalf("unexpected log body or missing local copy: %q", entry.Body)
		}
	}
	if transport.logs[0].Level != sentry.LogLevelInfo || transport.logs[1].Level != sentry.LogLevelError || transport.logs[2].Level != sentry.LogLevelError || !strings.Contains(transport.logs[2].Body, "status=500") {
		t.Fatalf("incorrect severity or HTTP status: %+v", transport.logs)
	}
	if !strings.Contains(local.String(), "CLI stays local") {
		t.Fatal("missing local CLI log")
	}
}
