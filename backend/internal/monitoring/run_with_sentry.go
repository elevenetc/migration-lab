package monitoring

import (
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
			if value != nil {
				hub.Recover(value)
			} else if err != nil {
				hub.CaptureException(err)
			}
			hub.Flush(2 * time.Second)
		}
		if value != nil {
			panic(value)
		}
	}()
	return run(hub)
}
