import { create } from 'zustand'
import { CreateTableMapEntry, Migration, fetchMigrations } from '../api/migrationApi'

interface MigrationState {
  migrations: Migration[]
  createTableMap: Record<string, CreateTableMapEntry>
  loading: boolean
  error: string | null
  loadMigrations: () => Promise<void>
}

export const useMigrationStore = create<MigrationState>((set) => ({
  migrations: [],
  createTableMap: {},
  loading: false,
  error: null,
  loadMigrations: async () => {
    set({ loading: true, error: null })
    try {
      const response = await fetchMigrations()
      set({
        migrations: response.timeline,
        createTableMap: response.createTableMap,
        loading: false
      })
    } catch (error) {
      set({ error: (error as Error).message, loading: false })
    }
  },
}))
