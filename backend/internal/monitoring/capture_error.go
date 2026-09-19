package monitoring

import (
	"context"
	"errors"

	"github.com/getsentry/sentry-go"
)

// CaptureError reports infrastructure errors even when an API returns them as
// a failed analysis result. CLI runs have no hub and remain uninstrumented.
func CaptureError(ctx context.Context, err error) {
	if err == nil || errors.Is(err, context.Canceled) {
		return
	}
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.CaptureException(err)
	}
}
