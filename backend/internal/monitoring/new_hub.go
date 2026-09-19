package monitoring

import (
	"errors"
	"strings"

	"github.com/getsentry/sentry-go"
)

// NewHub leaves reporting disabled unless a DSN was explicitly configured.
func NewHub(dsn, environment, release string) (*sentry.Hub, error) {
	if dsn == "" {
		return nil, nil
	}
	environment = strings.TrimSpace(environment)
	if environment == "" {
		return nil, errors.New("SENTRY_ENVIRONMENT must be set when SENTRY_DSN is configured")
	}
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn: dsn, Environment: environment, Release: release,
		AttachStacktrace: true,
		BeforeSend: func(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
			// Requests can contain local paths, SQL, cookies and credentials.
			if event.Request != nil {
				event.Request.QueryString = ""
				event.Request.Headers = nil
				event.Request.Cookies = ""
				event.Request.Data = ""
				event.Request.Env = nil
			}
			return event
		},
	})
	if err != nil {
		return nil, err
	}
	scope := sentry.NewScope()
	scope.SetTag("service", "backend")
	return sentry.NewHub(client, scope), nil
}
