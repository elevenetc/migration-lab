import { useEffect } from 'react'
import { Timeline } from './components/Timeline'
import { useMigrationStore } from './store/migrationStore'

function App() {
  const loadMigrations = useMigrationStore((state) => state.loadMigrations)
  const loadDatasets = useMigrationStore((state) => state.loadDatasets)
  const selectDataset = useMigrationStore((state) => state.selectDataset)
  const syncFromUrl = useMigrationStore((state) => state.syncFromUrl)
  const datasets = useMigrationStore((state) => state.datasets)
  const migrationId = useMigrationStore((state) => state.migrationId)
  const migrationsPath = useMigrationStore((state) => state.migrationsPath)
  const runMigrations = useMigrationStore((state) => state.runMigrations)
  const running = useMigrationStore((state) => state.running)
  const runResult = useMigrationStore((state) => state.runResult)

  useEffect(() => {
    loadMigrations()
    loadDatasets()
  }, [loadMigrations, loadDatasets])

  useEffect(() => {
    window.addEventListener('popstate', syncFromUrl)
    return () => window.removeEventListener('popstate', syncFromUrl)
  }, [syncFromUrl])

  return (
    <div style={{ width: '100vw', height: '100vh', display: 'flex', flexDirection: 'column' }}>
      <header style={{ padding: '16px', borderBottom: '1px solid #333', background: '#1a1a1a' }}>
        <h1 style={{ margin: 0, fontSize: '20px', color: '#fff' }}>Migration Lab</h1>
        <div style={{ marginTop: '8px', display: 'flex', alignItems: 'center', gap: '12px' }}>
          {datasets.length > 0 && (
            <select
              value={migrationsPath ? '' : migrationId ?? ''}
              onChange={(event) => selectDataset(event.target.value)}
              style={{
                padding: '8px 12px',
                background: '#2d3748',
                color: 'white',
                border: '1px solid #4a5568',
                borderRadius: '4px',
                cursor: 'pointer',
                fontSize: '14px',
              }}
            >
              <option value="" disabled>
                {migrationsPath ?? 'Select dataset'}
              </option>
              {datasets.map((dataset) => (
                <option key={dataset} value={dataset}>
                  {dataset}
                </option>
              ))}
            </select>
          )}
          <button
            onClick={runMigrations}
            disabled={running}
            style={{
              padding: '8px 16px',
              background: running ? '#4a5568' : '#22c55e',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: running ? 'not-allowed' : 'pointer',
            }}
          >
            {running ? 'Running...' : 'Run migrations'}
          </button>
          {runResult && (
            <span style={{ color: runResult.success ? '#22c55e' : '#ef4444', fontSize: '14px' }}>
              {runResult.success
                ? `Applied ${runResult.migrationsApplied} migrations`
                : runResult.message}
            </span>
          )}
        </div>
      </header>
      <main style={{ flex: 1, minHeight: 0 }}>
        <Timeline />
      </main>
    </div>
  )
}

export default App
