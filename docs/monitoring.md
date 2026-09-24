# Monitoring

Hosted Sentry captures frontend errors, backend panics, infrastructure failures and application
logs from both services. Reporting is disabled without a DSN. The CLI and exported HTML reports
do not send events or logs.

## Logs

Open **Logs** in Sentry and filter by environment and `service` (`frontend` or `backend`).
The existing DSNs and environment/release settings apply to logs too; no extra credentials are needed.

- Frontend: console messages are forwarded at their original level (`console.log` becomes `info`).
  A `Migration Lab frontend started` log is sent on initialization. For structured application logs,
  use `Sentry.logger.info('message', { key: 'value' })` (and the other severity methods).
- Backend: startup, HTTP requests, migration execution, seeding and cleanup logs are sent to Sentry
  and retained in local process output. Use `monitoring.Infof(ctx, ...)` or `monitoring.Errorf(ctx, ...)`
  for new application logs. Request logs and logs emitted inside a handler share the error event's
  `request_id`, `dataset_id` and `migration_id`. Request URLs use route templates without query strings.

Logs are batched, so allow a few seconds for delivery; the server flushes on shutdown. Plain Go
`log.Printf` calls and third-party container logs are not automatically forwarded. Console messages
and explicitly logged values are sent as written: keep SQL, local paths and credentials out of them.

## Local setup

Create Go and React projects in hosted Sentry. Copy their DSNs from **Settings → Projects → Client
Keys (DSN)** into the root `.env` (see [`.env.example`](../.env.example)):

```dotenv
SENTRY_DSN=<Go project DSN>
SENTRY_ENVIRONMENT=local
VITE_SENTRY_DSN=<React project DSN>
VITE_SENTRY_ENVIRONMENT=local
```

```bash
just dev-run
```

An environment is required when its app's DSN is set; missing values cause an initialization
error. Re-run after changing `.env`. For direct development, Vite reads `.env`; export
`SENTRY_DSN` and `SENTRY_ENVIRONMENT` before running `just run-backend`.

## Verify

Run in the app's browser console, then check the React project's **Issues** page (`local` environment):

```javascript
setTimeout(() => { throw new Error('Migration Lab Sentry test') }, 0)
```

Use the `request_id` tag to correlate frontend and backend HTTP failures. Expected migration
failures and HTTP 400/404 responses are not reported as Issues, but remain visible in logs.

For logs, reload the app and check **Logs** for `Migration Lab frontend started` and backend HTTP
request entries. You can also send a frontend test log from the browser console:

```javascript
console.info('Migration Lab Sentry log test')
```

After backend code changes, run `just dev-apply`; frontend source changes reload through Vite.

## Production

Set `SENTRY_ENVIRONMENT=demo` and `SENTRY_RELEASE=<git SHA>` for the backend. The frontend uses
`VITE_SENTRY_ENVIRONMENT` and `VITE_SENTRY_RELEASE` at **build time**. Set the same environment
for both apps; Compose passes these values through and maps the shared `SENTRY_RELEASE`.

Optional TypeScript source-map uploads require `SENTRY_ORG`, `SENTRY_PROJECT` (React project slug)
and `SENTRY_AUTH_TOKEN`. For Docker, export those plus `VITE_SENTRY_DSN` and `SENTRY_RELEASE`, then:

```bash
docker build -t migration-lab-frontend \
  --build-arg VITE_SENTRY_DSN \
  --build-arg VITE_SENTRY_ENVIRONMENT=demo \
  --build-arg VITE_SENTRY_RELEASE="$SENTRY_RELEASE" \
  --build-arg SENTRY_ORG --build-arg SENTRY_PROJECT \
  --secret id=sentry_auth_token,env=SENTRY_AUTH_TOKEN \
  frontend
```

Keep the auth token in build secrets, never in `VITE_*` variables. Uploaded maps are removed
from the build output.
