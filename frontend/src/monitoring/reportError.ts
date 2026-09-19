import * as Sentry from '@sentry/react'
import { ApiError } from '../api/apiError'

export function reportError(error: unknown, action: string, datasetId?: string | null, migrationId?: string) {
  if (window.__MIGRATION_DATA__ || !Sentry.isInitialized()) return
  // Invalid input and unknown datasets are normal API responses.
  if (error instanceof ApiError && error.status < 500) return
  if (error instanceof Error && error.name === 'AbortError') return

  Sentry.captureException(error, {
    tags: {
      action,
      dataset_id: datasetId ?? undefined,
      migration_id: migrationId,
      ...(error instanceof ApiError ? { http_status: error.status, request_id: error.requestId ?? undefined } : {}),
    },
  })
}
