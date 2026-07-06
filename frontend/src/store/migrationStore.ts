import { create } from 'zustand'
import { AnalysisResult, CreateTableMapEntry, Migration, fetchMigrations } from '../api/migrationApi'

interface MigrationState {
  migrations: Migration[]
  createTableMap: Record<string, CreateTableMapEntry>
  analysis: AnalysisResult | null
  loading: boolean
  error: string | null
  loadMigrations: () => Promise<void>
}

export const useMigrationStore = create<MigrationState>((set) => ({
  migrations: [],
  createTableMap: {},
  analysis: null,
  loading: false,
  error: null,
  loadMigrations: async () => {
    set({ loading: true, error: null })
    try {
      const response = await fetchMigrations()
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
}))
