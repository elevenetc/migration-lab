import { create } from 'zustand'
import { reportError } from '../monitoring/reportError'
import { CreateTableMapEntry, Migration, RunMigrationsResult, RuntimeAnalysisResult, StaticAnalysisResult, fetchDatasets, fetchMigrations, runMigrations, runRuntimeAnalysis } from '../api/migrationApi'

interface MigrationState {
  datasetId: string | null
  migrationsPath: string | null
  datasets: string[]
  migrations: Migration[]
  createTableMap: Record<string, CreateTableMapEntry>
  staticAnalysis: StaticAnalysisResult | null
  loading: boolean
  error: string | null
  running: boolean
  runResult: RunMigrationsResult | null
  runtimeRunning: boolean
  /** Migration the last runtime run measured alone, null when it measured every migration. */
  runtimeTarget: string | null
  runtimeAnalysis: RuntimeAnalysisResult | null
  runtimeError: string | null
  loadDatasets: () => Promise<void>
  loadMigrations: () => Promise<void>
  runMigrations: () => Promise<void>
  runRuntime: (migrationId: string | null) => Promise<void>
  closeRuntime: () => void
  selectDataset: (datasetId: string) => Promise<void>
  syncFromUrl: () => Promise<void>
}

function paramsFromUrl(): { datasetId: string | null; migrationsPath: string | null } {
  const searchParams = new URLSearchParams(window.location.search)
  return {
    datasetId: searchParams.get('datasetId'),
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
  staticAnalysis: null,
  loading: false,
  error: null,
  running: false,
  runResult: null,
  runtimeRunning: false,
  runtimeTarget: null,
  runtimeAnalysis: null,
  runtimeError: null,
  loadDatasets: async () => {
    if (isStaticReport()) return
    try {
      set({ datasets: await fetchDatasets() })
    } catch (error) {
      reportError(error, 'loadDatasets')
      set({ datasets: [] })
    }
  },
  selectDataset: async (datasetId: string) => {
    window.history.pushState(null, '', `?datasetId=${encodeURIComponent(datasetId)}`)
    set({ datasetId, migrationsPath: null, runResult: null, runtimeAnalysis: null, runtimeError: null })
    await get().loadMigrations()
  },
  syncFromUrl: async () => {
    const params = paramsFromUrl()
    if (params.datasetId === get().datasetId && params.migrationsPath === get().migrationsPath) return
    set({ ...params, runResult: null, runtimeAnalysis: null, runtimeError: null })
    await get().loadMigrations()
  },
  loadMigrations: async () => {
    set({ loading: true, error: null })
    try {
      const response = await fetchMigrations(get().datasetId, get().migrationsPath)
      set({
        migrations: response.timeline,
        createTableMap: response.createTableMap,
        staticAnalysis: response.analysis,
        loading: false
      })
    } catch (error) {
      reportError(error, 'loadMigrations', get().datasetId)
      set({
        error: (error as Error).message,
        loading: false,
        migrations: [],
        createTableMap: {},
        staticAnalysis: null,
      })
    }
  },
  runRuntime: async (migrationId: string | null) => {
    if (get().runtimeRunning) return
    set({ runtimeRunning: true, runtimeTarget: migrationId, runtimeAnalysis: null, runtimeError: null })
    try {
      const result = await runRuntimeAnalysis(get().datasetId, get().migrationsPath, migrationId)
      set({ runtimeRunning: false, runtimeAnalysis: result })
    } catch (error) {
      reportError(error, 'runRuntime', get().datasetId, migrationId ?? undefined)
      set({ runtimeRunning: false, runtimeError: (error as Error).message })
    }
  },
  closeRuntime: () => set({ runtimeAnalysis: null, runtimeError: null }),
  runMigrations: async () => {
    set({ running: true, runResult: null })
    try {
      const result = await runMigrations(get().datasetId)
      set({ running: false, runResult: result })
    } catch (error) {
      reportError(error, 'runMigrations', get().datasetId)
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
