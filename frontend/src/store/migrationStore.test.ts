import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import * as Sentry from '@sentry/react'

vi.mock('@sentry/react', () => ({ captureException: vi.fn(), isInitialized: () => true }))

beforeEach(() => {
  vi.clearAllMocks()
  vi.stubGlobal('window', { location: { search: '?datasetId=demo' } })
})
afterEach(() => vi.unstubAllGlobals())

describe('reporting caught API errors', () => {
  it('reports a server failure with its request ID and keeps the UI error', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('{}', { status: 500, headers: { 'X-Request-ID': 'request-123' } })))
    const { useMigrationStore } = await import('./migrationStore')
    await useMigrationStore.getState().loadMigrations()
    expect(Sentry.captureException).toHaveBeenCalledOnce()
    expect(Sentry.captureException).toHaveBeenCalledWith(expect.any(Error), {
      tags: expect.objectContaining({ action: 'loadMigrations', dataset_id: 'demo', request_id: 'request-123', http_status: 500 }),
    })
    expect(useMigrationStore.getState().error).toBe('Failed to fetch migrations')
  })

  it('reports swallowed dataset failures and network errors', async () => {
    const error = new TypeError('Failed to fetch')
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(error))
    const { useMigrationStore } = await import('./migrationStore')
    await useMigrationStore.getState().loadDatasets()
    expect(Sentry.captureException).toHaveBeenCalledWith(error, { tags: expect.objectContaining({ action: 'loadDatasets' }) })
    expect(useMigrationStore.getState().datasets).toEqual([])
  })

  it('reports caught runtime and run request failures', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('{}', { status: 503 })))
    const { useMigrationStore } = await import('./migrationStore')
    await useMigrationStore.getState().runRuntime('V2')
    expect(Sentry.captureException).toHaveBeenCalledWith(expect.any(Error), {
      tags: expect.objectContaining({ action: 'runRuntime', migration_id: 'V2', http_status: 503 }),
    })
    await useMigrationStore.getState().runMigrations()
    expect(Sentry.captureException).toHaveBeenCalledTimes(2)
    expect(useMigrationStore.getState().running).toBe(false)
  })

  it('does not report invalid input, unknown datasets or failed SQL results', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(new Response('{}', { status: 404 }))
      .mockResolvedValueOnce(new Response('{}', { status: 400 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ success: false, message: 'invalid SQL', migrationsApplied: 0 })))
    vi.stubGlobal('fetch', fetchMock)
    const { useMigrationStore } = await import('./migrationStore')
    await useMigrationStore.getState().loadMigrations()
    await useMigrationStore.getState().runRuntime('V2')
    await useMigrationStore.getState().runMigrations()
    expect(Sentry.captureException).not.toHaveBeenCalled()
    expect(useMigrationStore.getState().runResult?.success).toBe(false)
  })
})

describe('runtime analysis', () => {
  it('measures every migration when none is named and keeps the result', async () => {
    const analysis = { deadlineMs: 5000, rows: 1000, migrations: [], message: 'all 0 migrations completed in 0 ms' }
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify(analysis)))
    vi.stubGlobal('fetch', fetchMock)
    const { useMigrationStore } = await import('./migrationStore')

    await useMigrationStore.getState().runRuntime(null)

    expect(fetchMock).toHaveBeenCalledWith('/api/migrations/runtime-analysis?datasetId=demo', { method: 'POST' })
    expect(useMigrationStore.getState().runtimeRunning).toBe(false)
    expect(useMigrationStore.getState().runtimeTarget).toBeNull()
    expect(useMigrationStore.getState().runtimeAnalysis).toEqual(analysis)
  })

  it('names the migration when measuring one', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ migrations: [] })))
    vi.stubGlobal('fetch', fetchMock)
    const { useMigrationStore } = await import('./migrationStore')

    await useMigrationStore.getState().runRuntime('V2')

    expect(fetchMock).toHaveBeenCalledWith('/api/migrations/runtime-analysis?datasetId=demo&migration=V2', { method: 'POST' })
    expect(useMigrationStore.getState().runtimeTarget).toBe('V2')
  })

  it('starts no run while another is in flight', async () => {
    const fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    const { useMigrationStore } = await import('./migrationStore')

    useMigrationStore.setState({ runtimeRunning: true })
    await useMigrationStore.getState().runRuntime(null)

    expect(fetchMock).not.toHaveBeenCalled()
    useMigrationStore.setState({ runtimeRunning: false })
  })
})
