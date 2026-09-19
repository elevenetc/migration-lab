package monitoring

import "testing"

func TestNewHubRequiresExplicitDSN(t *testing.T) {
	t.Setenv("SENTRY_DSN", "https://public@example.invalid/1")
	hub, err := NewHub("", "", "")
	if err != nil || hub != nil {
		t.Fatalf("empty DSN should disable reporting: %v, %v", hub, err)
	}
}

func TestNewHubRejectsInvalidDSN(t *testing.T) {
	if _, err := NewHub("invalid", "local", ""); err == nil {
		t.Fatal("expected invalid DSN error")
	}
}

func TestNewHubRequiresEnvironmentWhenEnabled(t *testing.T) {
	for _, environment := range []string{"", " \t "} {
		hub, err := NewHub("https://public@example.invalid/1", environment, "")
		if hub != nil || err == nil || err.Error() != "SENTRY_ENVIRONMENT must be set when SENTRY_DSN is configured" {
			t.Fatalf("expected missing environment error, got hub %v, error %v", hub, err)
		}
	}
}

func TestNewHubUsesConfiguredEnvironment(t *testing.T) {
	hub, err := NewHub("https://public@example.invalid/1", "staging", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(hub.Client().Close)
	if got := hub.Client().Options().Environment; got != "staging" {
		t.Fatalf("expected staging environment, got %q", got)
	}
}
