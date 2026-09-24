package monitoring

import (
	"context"
	"log"

	"github.com/getsentry/sentry-go"
)

// Infof and Errorf keep local logs and send a copy to the request's Sentry hub.
// A context without a hub (including CLI runs) only writes locally.
func Infof(ctx context.Context, format string, args ...any) {
	logf(ctx, false, format, args...)
}

func Errorf(ctx context.Context, format string, args ...any) {
	logf(ctx, true, format, args...)
}

func logf(ctx context.Context, isError bool, format string, args ...any) {
	log.Printf(format, args...)
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		logger := sentry.NewLogger(ctx)
		entry := logger.Info()
		if isError {
			entry = logger.Error()
		}
		entry.Emitf(format, args...)
	}
}
