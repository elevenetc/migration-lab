package monitoring

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
)

// RunWithSentry captures startup errors and panics, then flushes before exit.
// Panics are rethrown after reporting so the process still fails normally.
func RunWithSentry(dsn, environment, release string, run func(*sentry.Hub) error) (err error) {
	hub, err := NewHub(dsn, environment, release)
	if err != nil {
		return err
	}
	defer func() {
		value := recover()
		if hub != nil {
			ctx := sentry.SetHubOnContext(context.Background(), hub)
			if value != nil {
				Errorf(ctx, "Server panicked: %v", value)
				hub.Recover(value)
			} else if err != nil {
				Errorf(ctx, "Server failed: %v", err)
				hub.CaptureException(err)
			}
			hub.Flush(2 * time.Second)
			hub.Client().Close()
		}
		if value != nil {
			panic(value)
		}
	}()
	return run(hub)
}
