import * as Sentry from '@sentry/react'

export function initSentry() {
  if (!import.meta.env.VITE_SENTRY_DSN || window.__MIGRATION_DATA__) return

  const environment = import.meta.env.VITE_SENTRY_ENVIRONMENT?.trim()
  if (!environment) {
    throw new Error('VITE_SENTRY_ENVIRONMENT must be set when VITE_SENTRY_DSN is configured')
  }

  Sentry.init({
    dsn: import.meta.env.VITE_SENTRY_DSN,
    environment,
    release: import.meta.env.VITE_SENTRY_RELEASE || undefined,
    dataCollection: {
      userInfo: false,
      cookies: false,
      httpHeaders: false,
      httpBodies: [],
      urlQueryParams: false,
      genAI: { inputs: false, outputs: false },
      databaseQueryData: false,
    },
    initialScope: { tags: { service: 'frontend' } },
    // Keep useful request/navigation breadcrumbs without collecting console
    // messages or DOM text, which may contain migration SQL and local paths.
    integrations: [Sentry.breadcrumbsIntegration({ console: false, dom: false })],
    beforeBreadcrumb(breadcrumb) {
      if (breadcrumb.data) {
        for (const key of ['url', 'from', 'to']) {
          const value = breadcrumb.data[key]
          if (typeof value === 'string') breadcrumb.data[key] = value.split(/[?#]/)[0]
        }
      }
      return breadcrumb
    },
    beforeSend(event) {
      if (event.request) {
        event.request = { url: event.request.url?.split(/[?#]/)[0], method: event.request.method }
      }
      return event
    },
  })
}
