import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import * as Sentry from '@sentry/react'
import { initSentry } from './initSentry'

vi.mock('@sentry/react', () => ({ init: vi.fn(), breadcrumbsIntegration: vi.fn() }))

beforeEach(() => {
  vi.clearAllMocks()
  vi.stubGlobal('window', {})
  vi.stubEnv('VITE_SENTRY_DSN', 'https://public@example.invalid/1')
  vi.stubEnv('VITE_SENTRY_ENVIRONMENT', 'local')
})
afterEach(() => { vi.unstubAllGlobals(); vi.unstubAllEnvs() })

it('disables reporting without a DSN and for offline reports', () => {
  vi.stubEnv('VITE_SENTRY_ENVIRONMENT', undefined)
  vi.stubEnv('VITE_SENTRY_DSN', '')
  initSentry()
  expect(Sentry.init).not.toHaveBeenCalled()
  vi.stubEnv('VITE_SENTRY_DSN', 'https://public@example.invalid/1')
  vi.stubGlobal('window', { __MIGRATION_DATA__: {} })
  initSentry()
  expect(Sentry.init).not.toHaveBeenCalled()
})

it.each([undefined, '', ' \t '])('rejects a missing or blank environment (%s) when enabled', (environment) => {
  vi.stubEnv('VITE_SENTRY_ENVIRONMENT', environment)
  expect(initSentry).toThrow('VITE_SENTRY_ENVIRONMENT must be set when VITE_SENTRY_DSN is configured')
  expect(Sentry.init).not.toHaveBeenCalled()
})

it('uses the configured environment', () => {
  vi.stubEnv('VITE_SENTRY_ENVIRONMENT', 'staging')
  initSentry()
  expect(Sentry.init).toHaveBeenCalledWith(expect.objectContaining({ environment: 'staging' }))
})

it('keeps sensitive data collection disabled after the SDK resolves its defaults', async () => {
  initSentry()
  const options = vi.mocked(Sentry.init).mock.calls[0][0]!
  const { BrowserClient } = await vi.importActual<typeof Sentry>('@sentry/react')
  const client = new BrowserClient({
    ...options,
    integrations: [],
    stackParser: () => [],
    transport: () => ({ send: async () => ({}), flush: async () => true }),
  })
  expect(client.getDataCollectionOptions()).toMatchObject({
    userInfo: false,
    cookies: false,
    httpHeaders: { request: false, response: false },
    httpBodies: [],
    urlQueryParams: false,
    genAI: { inputs: false, outputs: false },
    databaseQueryData: false,
  })
  await client.close()
})

it('removes URL queries and request content from events and breadcrumbs', async () => {
  initSentry()
  const options = vi.mocked(Sentry.init).mock.calls[0][0]!
  const event = await options.beforeSend!({ type: undefined, request: { url: 'http://localhost/?migrationsPath=/private/sql#secret', headers: { Authorization: 'secret' }, data: 'SQL' } }, {})
  expect(event?.request).toEqual({ url: 'http://localhost/', method: undefined })
  expect(options.beforeBreadcrumb!({ category: 'fetch', data: { url: '/api/migrations?migrationsPath=/private/sql', status_code: 500 } })).toEqual({
    category: 'fetch', data: { url: '/api/migrations', status_code: 500 },
  })
})
