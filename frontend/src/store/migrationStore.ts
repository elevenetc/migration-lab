import { create } from 'zustand'
import { AnalysisResult, CreateTableMapEntry, Migration, RunMigrationsResult, RuntimeResult, fetchDatasets, fetchMigrations, runMigrations, runRuntimeAnalysis } from '../api/migrationApi'

interface MigrationState {
  migrationId: string | null
  migrationsPath: string | null
  datasets: string[]
  migrations: Migration[]
  createTableMap: Record<string, CreateTableMapEntry>
  analysis: AnalysisResult | null
  loading: boolean
  error: string | null
  running: boolean
  runResult: RunMigrationsResult | null
  /** Migration currently being measured, null when no run is in flight. */
  runtimeTarget: string | null
  runtimeResult: RuntimeResult | null
  runtimeError: string | null
  loadDatasets: () => Promise<void>
  loadMigrations: () => Promise<void>
  runMigrations: () => Promise<void>
  runRuntime: (migrationId: string) => Promise<void>
  closeRuntime: () => void
  selectDataset: (migrationId: string) => Promise<void>
  syncFromUrl: () => Promise<void>
}

function paramsFromUrl(): { migrationId: string | null; migrationsPath: string | null } {
  const searchParams = new URLSearchParams(window.location.search)
  return {
    migrationId: searchParams.get('migrationId'),
    migrationsPath: searchParams.get('migrationsPath'),
  }
}

// Static reports embed their data and have no backend to query for datasets.
const isStaticReport = () => Boolean(window.__MIGRATION_DATA__)

export const useMigrationStore = create<MigrationState>((set, get) => ({
  ...paramsFromUrl(),
  datasets: [],
  migrations: [],
  createTableMap: {},
  analysis: null,
  loading: false,
  error: null,
  running: false,
  runResult: null,
  runtimeTarget: null,
  runtimeResult: null,
  runtimeError: null,
  loadDatasets: async () => {
    if (isStaticReport()) return
    try {
      set({ datasets: await fetchDatasets() })
    } catch {
      set({ datasets: [] })
    }
  },
  selectDataset: async (migrationId: string) => {
    window.history.pushState(null, '', `?migrationId=${encodeURIComponent(migrationId)}`)
    set({ migrationId, migrationsPath: null, runResult: null, runtimeResult: null, runtimeError: null })
    await get().loadMigrations()
  },
  syncFromUrl: async () => {
    const params = paramsFromUrl()
    if (params.migrationId === get().migrationId && params.migrationsPath === get().migrationsPath) return
    set({ ...params, runResult: null, runtimeResult: null, runtimeError: null })
    await get().loadMigrations()
  },
  loadMigrations: async () => {
    set({ loading: true, error: null })
    try {
      const response = await fetchMigrations(get().migrationId, get().migrationsPath)
      set({
        migrations: response.timeline,
        createTableMap: response.createTableMap,
        analysis: response.analysis,
        loading: false
      })
    } catch (error) {
      set({
        error: (error as Error).message,
        loading: false,
        migrations: [],
        createTableMap: {},
        analysis: null,
      })
    }
  },
  runRuntime: async (migrationId: string) => {
    if (get().runtimeTarget) return
    set({ runtimeTarget: migrationId, runtimeResult: null, runtimeError: null })
    try {
      const result = await runRuntimeAnalysis(get().migrationId, get().migrationsPath, migrationId)
      set({ runtimeTarget: null, runtimeResult: result })
    } catch (error) {
      set({ runtimeTarget: null, runtimeError: (error as Error).message })
    }
  },
  closeRuntime: () => set({ runtimeResult: null, runtimeError: null }),
  runMigrations: async () => {
    set({ running: true, runResult: null })
    try {
      const result = await runMigrations(get().migrationId)
      set({ running: false, runResult: result })
    } catch (error) {
      set({
        running: false,
        runResult: {
          success: false,
          message: (error as Error).message,
          migrationsApplied: 0
        }
      })
    }
  },
}))
