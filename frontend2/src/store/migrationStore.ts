import { create } from 'zustand'
import { AnalysisResult, CreateTableMapEntry, Migration, RunMigrationsResult, fetchMigrations, runMigrations } from '../api/migrationApi'

interface MigrationState {
  migrationId: string | null
  migrations: Migration[]
  createTableMap: Record<string, CreateTableMapEntry>
  analysis: AnalysisResult | null
  loading: boolean
  error: string | null
  running: boolean
  runResult: RunMigrationsResult | null
  loadMigrations: () => Promise<void>
  runMigrations: () => Promise<void>
}

const migrationId = new URLSearchParams(window.location.search).get('migrationId')

export const useMigrationStore = create<MigrationState>((set, get) => ({
  migrationId,
  migrations: [],
  createTableMap: {},
  analysis: null,
  loading: false,
  error: null,
  running: false,
  runResult: null,
  loadMigrations: async () => {
    set({ loading: true, error: null })
    try {
      const response = await fetchMigrations(get().migrationId)
      set({
        migrations: response.timeline,
        createTableMap: response.createTableMap,
        analysis: response.analysis,
        loading: false
      })
    } catch (error) {
      set({ error: (error as Error).message, loading: false })
    }
  },
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
